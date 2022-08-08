package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sort"
	"time"
	"strings"

	//"github.com/gleanerio/gleaner/internal/common"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/viper"
)

var viperVal, modeVal string
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
	// flag.StringVar(&sourceVal, "source", "", "Override config file source")
	flag.BoolVar(&setupVal, "setup", false, "Run Gleaner configuration check and exit")
	flag.StringVar(&viperVal, "cfg", "config", "Configuration file")
	flag.StringVar(&modeVal, "mode", "mode", "Set the mode")
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

	if isFlagPassed("mode") {
		m := v1.GetStringMap("summoner")
		m["mode"] = modeVal
		v1.Set("summoner", m)
	}

	// Set up the minio connector
	mc := MinioConnection(v1)

	// -------------- list objects
	//lo := listObjDates(mc)
	//for i := range lo {
	//	fmt.Printf("%s :: %s \n", lo[i].Name, lo[i].Date)
	//}

	//* copy summoned to the archive
	//* empty summoned and milled

	//  Move == copy and remove
	copy(mc,  "gleaner.oihqueue","summoned/maspawio","archive/maspawio" )
	remove(mc,  "gleaner.oihqueue","summoned/maspawio" )
}



func copy(mc *minio.Client, b, s, d string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	objectCh := mc.ListObjects(ctx, b, minio.ListObjectsOptions{
		Prefix:    s,
		Recursive: true,
	})
	for object := range objectCh {
		if object.Err != nil {
			fmt.Println(object.Err)
		}

		// get the object name
		n := object.Key[strings.LastIndex(object.Key, "/")+1:]

		fmt.Printf("%s/%s\n", s, n)

		srcOpts := minio.CopySrcOptions{
		Bucket: b,
		Object: fmt.Sprintf("%s/%s", s, n),
		}

		// Destination object``
		dstOpts := minio.CopyDestOptions{
		Bucket: b,
		Object: fmt.Sprintf("%s/%s", d, n),
		}

		// Copy object call
		_, err := mc.CopyObject(context.Background(), dstOpts, srcOpts)
		if err != nil {
		fmt.Println(err)
		return
		}

	}
}

func remove(mc *minio.Client, s, p string) {
	//objectsCh := make(chan string)
	objectsCh := make(chan minio.ObjectInfo)

	// Send object names that are needed to be removed to objectsCh
	go func() {
		defer close(objectsCh)
		// List all objects from a bucket-name with a matching prefix.
		opts := minio.ListObjectsOptions{
			Recursive: true,
			Prefix:    p,
		}
		for object := range mc.ListObjects(context.Background(), s, opts) {
			if object.Err != nil {
				log.Fatal(object.Err)
			}
			//objectsCh <- object.Key
			objectsCh <- object
		}

	}()

	opts := minio.RemoveObjectsOptions{
		GovernanceBypass: true,
	}

	// remove all the objects in the channel
	for rErr := range mc.RemoveObjects(context.Background(), s, objectsCh, opts) {
		fmt.Println("Error detected during deletion: ", rErr)
	}

	// Also remove the bucket  (not active at this time, no real need)
	//err := mc.RemoveBucket(context.Background(), b)
	//if err != nil {
	//fmt.Println(err)
	//return
	//}

}

// ListObjects  return the list of objects from the
// bucket sorted by date with newest to the end
func listObjDates(mc *minio.Client) []BucketObjects {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var bo []BucketObjects

	objectCh := mc.ListObjects(ctx, "gleaner.oihqueue", minio.ListObjectsOptions{
		Prefix:    "summoned/maspawio",
		Recursive: true,
	})
	for object := range objectCh {
		if object.Err != nil {
			fmt.Println(object.Err)
			return bo
		}

		// grab object metadata
		objInfo, err := mc.StatObject(context.Background(), "gleaner.oihqueue", object.Key, minio.StatObjectOptions{})
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
