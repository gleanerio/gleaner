package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	awscred "github.com/aws/aws-sdk-go/aws/credentials"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/earthcubearchitecture-project418/gleaner/internal/common"
	"github.com/earthcubearchitecture-project418/gleaner/internal/organizations"
	"github.com/earthcubearchitecture-project418/gleaner/internal/summoner/acquire"
	"github.com/knakk/rdf"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/viper"
	"github.com/xitongsys/parquet-go-source/s3"
	"github.com/xitongsys/parquet-go/writer"
)

var viperVal string
var setupVal bool

type BucketObjects struct {
	Name string
	Date time.Time
}

type Qset struct {
	Subject   string `parquet:"name=Subject,  type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Predicate string `parquet:"name=Predicate,  type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Object    string `parquet:"name=Object,  type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Graph     string `parquet:"name=Graph,  type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
}

func init() {
	log.SetFlags(log.Lshortfile)
	// log.SetOutput(ioutil.Discard) // turn off all logging

	flag.BoolVar(&setupVal, "setup", false, "Run Gleaner configuration check and exit")
	flag.StringVar(&viperVal, "cfg", "config", "Configuration file")
}

// Simple test of the nanoprov function
func main() {
	flag.Parse() // parse any command line flags...

	var v1 *viper.Viper
	var err error

	// Load the config file and set some defaults (config overrides)
	if isFlagPassed("cfg") {
		v1, err = readConfig(viperVal, map[string]interface{}{
			"sqlfile": "",
			"bucket":  "",
			"minio": map[string]string{
				"address":   "localhost",
				"port":      "9000",
				"accesskey": "",
				"secretkey": "",
			},
		})
		if err != nil {
			panic(fmt.Errorf("error when reading config: %v", err))
		}
	}

	// Set up the minio connector
	mc := MinioConnection(v1)

	// func StoreProv(k, sha, urlloc string, mc *minio.Client) error {
	// err = acquire.StoreProv(v1, mc, k, sha, urlloc)

	// -------------- Make org rdf
	organizations.BuildGraph(mc, v1)
	if err != nil {
		log.Println("Org build failed")
		log.Println(err)
	}

	// -------------- list objects
	// lo := ListObjects(mc)
	// for i := range lo {
	// 	fmt.Printf("%s :: %s \n", lo[i].Name, lo[i].Date)
	// }

	// -------------- get the graph
	// GetGraph(v1)

	// -------------- parquet builder
	//err = Bkt2File(v1, "name", "gleaner", "prov/samplesearth", mc)
	//if err != nil {
	//log.Print(err)
	//}

	//foo := []byte("_:b0 <http://schema.org/jobTitle> 'Professor' .")
	//z := bytes.NewBuffer(foo)

	//err = PrqtRDFToS3(v1, "gleaner", "results/test.parquet", "us-east-1", z)
	//if err != nil {
	//log.Print(err)
	//}
}

func PrqtRDFToS3(v1 *viper.Viper, bucket, key, region string, rbb *bytes.Buffer) error {
	mcfg := v1.Sub("minio")
	//endpoint := fmt.Sprintf("%s:%s", mcfg.GetString("address"), mcfg.GetString("port"))
	accessKeyID := mcfg.GetString("accesskey")
	secretAccessKey := mcfg.GetString("secretkey")
	//useSSL := mcfg.GetBool("ssl")

	// Make a parquet file
	ctx := context.Background()

	// aws.Config{}
	// // LOCAL create new S3 file writer
	fw, err := s3.NewS3FileWriter(ctx, bucket, key, nil, &aws.Config{Region: aws.String(region),
		Endpoint:         aws.String("http://192.168.86.45:32773"),
		Credentials:      awscred.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true)})
	if err != nil {
		log.Println("Can't create s3 file writer", err)
		return err
	}

	// TODO..  look at doing a memory file?   and then load to S3 and delete and move on?
	// https://github.com/xitongsys/parquet-go-source/blob/master/examples/memfs_write.go

	// fw, err := s3.NewS3FileWriter(ctx, bucket, key, nil, &aws.Config{Region: aws.String(region)})
	// if err != nil {
	// 	log.Println("Can't open file", err)
	// 	return err
	// }

	// set up parquet file
	pw, err := writer.NewParquetWriter(fw, new(Qset), 4)
	if err != nil {
		log.Println("Can't create parquet writer", err)
		return err
	}

	pw.RowGroupSize = 128 * 1024 * 1024 //128M
	pw.PageSize = 8 * 1024              //8K
	// pw.CompressionType = parquet.CompressionCodec_SNAPPY

	// scanner := bufio.NewScanner(strings.NewReader(r))
	scanner := bufio.NewScanner(bytes.NewReader(rbb.Bytes()))
	for scanner.Scan() {
		rdfb := bytes.NewBufferString(scanner.Text())
		dec := rdf.NewQuadDecoder(rdfb, rdf.NQuads)

		spog, err := dec.Decode()
		if err != nil {
			log.Println("can't decode", err)
			return err
		}

		qs := Qset{Subject: spog.Subj.String(), Predicate: spog.Pred.String(), Object: spog.Obj.String(), Graph: spog.Ctx.String()}

		// log.Print(qs)

		if err = pw.Write(qs); err != nil {
			log.Println("Write error", err)
			return err
		}

	}
	if err := scanner.Err(); err != nil {
		log.Println("Error during scan")
		log.Println(err)
		return err
	}

	pw.Flush(true)

	if err = pw.WriteStop(); err != nil {
		log.Println("WriteStop error", err)
		return err
	}

	err = fw.Close()
	if err != nil {
		log.Println(err)
		log.Println("Error closing S3 file writer")
		return err
	}

	return err
}

func Bkt2File(v1 *viper.Viper, name, bucket, prefix string, mc *minio.Client) error {
	log.Println("Start pipe reader / writer sequence")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pr, pw := io.Pipe()     // TeeReader of use?
	lwg := sync.WaitGroup{} // work group for the pipe writes...
	lwg.Add(2)

	// params for list objects calls
	doneCh := make(chan struct{}) // , N) Create a done channel to control 'ListObjectsV2' go routine.
	defer close(doneCh)           // Indicate to our routine to exit cleanly upon return.
	// isRecursive := true

	go func() {
		defer lwg.Done()
		defer pw.Close()
		for object := range mc.ListObjects(ctx, "gleaner", minio.ListObjectsOptions{
			Prefix:    "summoned/samplesearth",
			Recursive: true,
		}) {
			fo, err := mc.GetObject(ctx, bucket, object.Key, minio.GetObjectOptions{})
			if err != nil {
				fmt.Println(err)
			}

			sb := new(strings.Builder)
			_, err = io.Copy(sb, fo)
			if err != nil {
				fmt.Println(err)
			}

			proc, options := common.JLDProc(v1)
			nq, err := common.JLD2nq(sb.String(), proc, options)
			if err != nil {
				fmt.Println(err)
			}

			pw.Write([]byte(nq))
		}
	}()

	// log.Printf("%s_graph.nq", name)

	// go function to write to minio from pipe
	// go func() {
	// 	defer lwg.Done()
	// 	_, err := mc.PutObject("gleaner", name, pr, -1, minio.PutObjectOptions{})
	// 	if err != nil {
	// 		log.Println(err)
	// 	}
	// }()

	// Note: We can also make a file and pipe write to that, keep this code around in case
	f, err := os.Create(fmt.Sprintf("%s_graph.nq", "test")) // prefix)) // needs a f.Close() later
	if err != nil {
		log.Println(err)
	}
	// go function to write to file from pipe
	go func() {
		defer lwg.Done()
		if _, err := io.Copy(f, pr); err != nil {
			log.Fatal(err)
		}
	}()

	lwg.Wait() // wait for the pipe read writes to finish
	pw.Close()
	pr.Close()

	f.Close()

	return nil
}

// GetGraph
// modify config file with sitegraph entry
// download URL
// load to minio
// generate prov
// generate org
func GetGraph(v1 *viper.Viper) string {
	// read graph info from v1
	var domains []acquire.Sources
	err := v1.UnmarshalKey("graphs", &domains)
	if err != nil {
		log.Println(err)
	}

	for k := range domains {
		fmt.Println(domains[k].URL)

		d, err := getJSON(domains[k].URL)
		if err != nil {
			fmt.Println("error with reading graph JSON")
		}

		// load graph
		// build prov?

		fmt.Println(len(d))
	}

	return ""
}

func getJSON(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("GET error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Status error: %v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Read body: %v", err)
	}

	return string(data), nil
}

// ListObjects  return the list of objects from the
// bucket sorted by date with newest to the end
func ListObjects(mc *minio.Client) []BucketObjects {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var bo []BucketObjects

	objectCh := mc.ListObjects(ctx, "gleaner", minio.ListObjectsOptions{
		Prefix:    "summoned/samplesearth",
		Recursive: true,
	})
	for object := range objectCh {
		if object.Err != nil {
			fmt.Println(object.Err)
			return bo
		}

		// grab object metadata
		objInfo, err := mc.StatObject(context.Background(), "gleaner", object.Key, minio.StatObjectOptions{})
		if err != nil {
			fmt.Println(err)
			return bo
		}

		o := BucketObjects{Name: object.Key, Date: objInfo.LastModified}
		//fmt.Println(object.Key)
		bo = append(bo, o)
	}

	sort.Slice(bo, func(i, j int) bool { return bo[i].Date.Before(bo[j].Date) })

	return bo
}

func MinioConnection(v1 *viper.Viper) *minio.Client {
	//mcfg := v1.GetStringMapString("minio")
	mcfg := v1.Sub("minio")
	endpoint := fmt.Sprintf("%s:%s", mcfg.GetString("address"), mcfg.GetString("port"))
	accessKeyID := mcfg.GetString("accesskey")
	secretAccessKey := mcfg.GetString("secretkey")
	useSSL := mcfg.GetBool("ssl")

	minioClient, err := minio.New(endpoint,
		&minio.Options{Creds: credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
			Secure: useSSL})
	if err != nil {
		log.Fatalln(err)
	}

	return minioClient
}

func readConfig(filename string, defaults map[string]interface{}) (*viper.Viper, error) {
	v := viper.New()
	for key, value := range defaults {
		v.SetDefault(key, value)
	}
	v.SetConfigName(filename)
	v.AddConfigPath(".")
	v.AutomaticEnv()
	err := v.ReadInConfig()
	return v, err
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

//func PrqtRDFToS3(bucket, key, region string, rbb *bytes.Buffer) error {
//ctx := context.Background()

//fw, err := s3.NewS3FileWriter(ctx, bucket, key, nil, &aws.Config{Region: aws.String(region)})
//if err != nil {
//log.Println("Can't open file", err)
//return err
//}

//// set up parquet file
//pw, err := writer.NewParquetWriter(fw, new(Qset), 4)
//if err != nil {
//log.Println("Can't create parquet writer", err)
//return err
//}

//pw.RowGroupSize = 128 * 1024 * 1024 //128M
//pw.PageSize = 8 * 1024              //8K
//pw.CompressionType = parquet.CompressionCodec_SNAPPY

//// scanner := bufio.NewScanner(strings.NewReader(r))
//scanner := bufio.NewScanner(bytes.NewReader(rbb.Bytes()))
//for scanner.Scan() {
//rdfb := bytes.NewBufferString(scanner.Text())
//dec := rdf.NewQuadDecoder(rdfb, rdf.NQuads)

//spog, err := dec.Decode()
//if err != nil {
//log.Println("can't decode", err)
//return err
//}

//qs := Qset{Subject: spog.Subj.String(), Predicate: spog.Pred.String(), Object: spog.Obj.String(), Graph: spog.Ctx.String()}

//// log.Print(qs)

//if err = pw.Write(qs); err != nil {
//log.Println("Write error", err)
//return err
//}

//}
//if err := scanner.Err(); err != nil {
//log.Println("Error during scan")
//log.Println(err)
//return err
//}

//pw.Flush(true)

//if err = pw.WriteStop(); err != nil {
//log.Println("WriteStop error", err)
//return err
//}

//err = fw.Close()
//if err != nil {
//log.Println(err)
//log.Println("Error closing S3 file writer")
//return err
//}

//return err
//}
