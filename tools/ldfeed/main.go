package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/earthcubearchitecture-project418/gleaner/internal/common"
	"github.com/earthcubearchitecture-project418/gleaner/internal/objects"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/viper"
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

	// -------------- get the graph
	fn, err := GetFeed(mc, v1)
	if err != nil {
		log.Print(err)
	}

	log.Println(fn)
}

func GetFeed(mc *minio.Client, v1 *viper.Viper) (string, error) {
	// read graph info from v1
	log.Println("feedgraph indexing")

	var domains []objects.Sources
	err := v1.UnmarshalKey("feedgraphs", &domains)
	if err != nil {
		log.Println(err)
	}

	for k := range domains {
		fmt.Println(domains[k].URL)

		d, err := getJSON(domains[k].URL)
		if err != nil {
			fmt.Println("error with reading graph JSON")
		}

		fd, err := framedoc(v1, []byte(d))
		if err != nil {
			fmt.Println(" error ....")
		}

		log.Println(fd)

		//st := time.Now()
		//log.Printf("Hash start time: %s \n", st)

		//sha := common.GetSHA(d) // Don't normalize big files..

		//et := time.Now()
		//diff := et.Sub(st)
		//log.Printf("Hash end time: %s \n", et)
		//log.Printf("Hash run time: %f \n", diff.Minutes())

		//// Upload the file
		//objectName := fmt.Sprintf("summoned/%s/%s.jsonld", domains[k].Name, sha)
		//_, err = graph.LoadToMinio(d, "gleaner", objectName, mc)
		//if err != nil {
		//return objectName, err
		//}

		//// build prov?
		//err = acquire.StoreProv(v1, mc, domains[k].Name, sha, domains[k].URL)
		//if err != nil {
		//log.Println(err)
		//}

		fmt.Println(len(d))
	}

	return "test", err
}

func framedoc(v1 *viper.Viper, b []byte) (string, error) {
	proc, options := common.JLDProc(v1)

	frame := map[string]interface{}{
		"@context":            map[string]interface{}{"@vocab": "http://schema.org/", "dcterms": "http://purl.org/dc/terms/"},
		"@explicit":           true,
		"dcterms:isVersionOf": "",
		"dcterms:modified":    "",
	}

	// "@context": {"schema": "https://schema.org/",
	// "dcterms": "http://purl.org/dc/terms/"},
	// "@explicit": "true",
	// "dcterms:isVersionOf": "",
	// "dcterms:modified": ""

	var myInterface interface{}
	err := json.Unmarshal(b, &myInterface)
	if err != nil {
		log.Println("Error when transforming JSON-LD document to interface:", err)
	}

	framedDoc, err := proc.Frame(myInterface, frame, options) // do I need the options set in order to avoid the large context that seems to be generated?
	if err != nil {
		log.Println("Error when trying to frame document", err)
	}

	graph := framedDoc["@graph"]

	jsonm, err := json.MarshalIndent(graph, "", " ")
	if err != nil {
		log.Println("Error trying to marshal data", err)
	}

	log.Println(string(jsonm))

	//df := []DataFrame{}
	//json.Unmarshal(jsonm, &df)

	//if len(df) > 0 {
	//return df[0].Description, nil

	//}

	return "", nil
}

func getJSON(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("GET error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status error: %v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %v", err)
	}

	return string(data), nil
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
