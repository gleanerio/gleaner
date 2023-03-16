package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/gleanerio/gleaner/internal/objects"
	"github.com/gleanerio/gleaner/internal/summoner/acquire"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/viper"
	bolt "go.etcd.io/bbolt"
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

	// read config file
	miniocfg := v1.GetStringMapString("minio")
	bucketName := miniocfg["bucket"] //   get the top level bucket for all of gleaner operations from config file
	// setup the KV store to hold a record of indexed resources
	db, err := bolt.Open("gleaner.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	ru, err := acquire.ResourceURLs(v1, mc, false, db)
	// hru := acquire.ResourceURLs(v1, true)
	// log.Println(len(ru["samplesearth"].URL))
	// log.Println(len(hru["samplesearth"].URL))

	d := "oceanexperts" // domain to test  obis marinetraining oceanexperts

	var u []string
	for k := range ru[d] {
		u = append(u, ru[d][k])
	}

	fmt.Println("print sitemap urls -----------------------------------------------")
	for k := range u {
		fmt.Println(u[k])
	}

	// // TEST
	// u = append(u, "this is a test")

	// s3select prov call
	oa := objects.ProvURLs(v1, mc, bucketName, fmt.Sprintf("prov/%s", d))
	fmt.Println("print s3select urls -----------------------------------------------")
	for k := range oa {
		fmt.Println(oa[k])
	}

	diff := difference(u, oa)

	fmt.Printf("Len sitemap: %d   Len prov: %d   Len diff: %d \n", len(u), len(oa), len(diff))

	fmt.Println("print diff urls -----------------------------------------------")
	for k := range diff {
		fmt.Println(diff[k])
	}

}

// difference returns the elements in `a` that aren't in `b`.
func difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
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
