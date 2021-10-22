package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/earthcubearchitecture-project418/gleaner/internal/config"
	"github.com/gocarina/gocsv"
	"github.com/spf13/viper"
	"github.com/yunabe/easycsv"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

//need to add cobra
// https://github.com/spf13/cobra/blob/master/user_guide.md#using-the-cobra-generator

type Source = config.Sources
type SourceConfig = config.SourcesConfig
type MinoConfig = config.Minio

var minioVal, portVal, accessVal, secretVal, bucketVal, cfgVal, viperVal, cfgPath string
var glrVal, nabuVal, sourcesVal, templateGleaner, templateNabu string
var sslVal bool
var check, cfginit, cfgval, cfgload, cfgrun, cfggen bool

var configFullPath string

var configBaseFiles = map[string]string{"gleaner": "gleaner_base.yaml", "sources": "sources.csv",
	"nabu": "nabu_base.yaml", "minio": "minio.yaml", "readme": "readme.txt"}
var gleanerBaseName = "gleaner"
var nabuBaseName = "nabu"

var (
	gleanerTemplate = map[string]interface{}{
		"minio": map[string]string{
			"address":   "localhost",
			"port":      "9000",
			"accesskey": "",
			"secretkey": "",
		},
		"gleaner":     "",
		"context":     "",
		"contextmaps": "",
		"summoner":    "",
		"millers":     ",",
		"sources":     "",
	}
	//minioTemplate = map[string]interface{}{
	//	"minio": map[string]string{
	//		"address":   "localhost",
	//		"port":      "9000",
	//		"accesskey": "",
	//		"secretkey": "",
	//	},
	//
	//}
	minioTemplate = map[string]interface{}{}
	nabuTemplate  = map[string]interface{}{
		"minio":   "",
		"sparql":  "",
		"objects": "",
	}
)

func init() {
	akey := os.Getenv("MINIO_ACCESS_KEY")
	skey := os.Getenv("MINIO_SECRET_KEY")

	flag.StringVar(&minioVal, "address", "localhost", "FQDN for server")
	flag.StringVar(&portVal, "port", "9000", "Port for minio server, default 9000")
	flag.StringVar(&accessVal, "access", akey, "Access Key ID")
	flag.StringVar(&secretVal, "secret", skey, "Secret access key")
	flag.StringVar(&bucketVal, "bucket", "gleaner", "The configuration bucket")

	flag.BoolVar(&sslVal, "ssl", false, "Use SSL boolean")

	flag.StringVar(&cfgVal, "cfg", "local", "Cofiguration Name")
	flag.StringVar(&cfgPath, "cfgPath", "configs", "Cofiguration path")
	flag.StringVar(&templateGleaner, "template", "template_v2.0", "Configuration Template or Cofiguration file")
	flag.StringVar(&templateNabu, "template_nabu", "template_nabu", "Configuration Template or Cofiguration file")
	flag.StringVar(&glrVal, "gleaner", "gleaner.yaml", "output gleaner file to")
	flag.StringVar(&nabuVal, "nabu", "nabu.yaml", "output nabu file to")
	flag.StringVar(&sourcesVal, "sourcemaps", "sources.csv", "sources as csv")

	flag.BoolVar(&check, "check", false, "Check the running minio setup")
	flag.BoolVar(&cfginit, "cfginit", false, "Init a config template")
	flag.BoolVar(&cfggen, "cfggen", false, "Generate or Regenerate gleaner and nabu configuration files")
	flag.BoolVar(&cfgval, "cfgval", false, "Validate a config file is parsable")

	// cfgload is NOT boolean, it's a string pointing to the config file to load and run
	flag.BoolVar(&cfgload, "cfgload", false, "Load the config into the mino input queue bucket")

	// cfgrun is NOT needed..  load will validate and load..  the buckets will trigger
	// the webhook call to gleaner to check the input queue
	flag.BoolVar(&cfgrun, "cfgrun", false, "Call Gleaner to check input queue bucket")
}

func main() {
	var n1 *viper.Viper // nabu
	var g1 *viper.Viper // gleaner
	var m1 *viper.Viper // minio
	var sources []Source
	var err error

	fmt.Println("EarthCube Gleaner Control")
	flag.Parse()

	//if isFlagPassed("cfg") {
	//	fmt.Println("Config name passed")
	//	configFullPath =
	//	//v1, err = readConfig(cfgVal, gleanerTemplate)
	//	//if err != nil {
	//	//	panic(fmt.Errorf("error when reading config: %v", err))
	//	//}
	//}

	// read the template, always
	//	t1, err = readConfig(templateGleaner, path.Join(cfgPath) gleanerTemplate)
	//		if err != nil {
	//			panic(fmt.Errorf("error when reading config: %v", err))
	//		}
	// init s3 client
	// init s3 buckets
	//err := initBucket()
	//if err != nil {
	//	fmt.Println(err)
	//}
	//if isFlagPassed("sourcemaps") {
	//	fmt.Println("sources passed")
	//	sources, err = readSourcesGoCSV(sourcesVal)
	//	if err != nil {
	//		panic(fmt.Errorf("error when reading sourcemap listing: %v", err))
	//	}
	//}

	// init config
	if cfginit {
		err := initCfg(cfgPath, cfgVal, configBaseFiles)
		if err != nil {
			fmt.Println(err)
		}
	}
	// read files
	if cfgval || cfggen {
		//read the template, always
		g1, err = readConfig(fileNameWithoutExtTrimSuffix(configBaseFiles["gleaner"]), path.Join(cfgPath, cfgVal), gleanerTemplate)
		if err != nil {
			panic(fmt.Errorf("error when reading gleaner config: %v", err))
		}
		n1, err = readConfig(fileNameWithoutExtTrimSuffix(configBaseFiles["nabu"]), path.Join(cfgPath, cfgVal), nabuTemplate)
		if err != nil {
			panic(fmt.Errorf("error when reading nabu config: %v", err))
		}
		m1, err = readConfigNoDefault(fileNameWithoutExtTrimSuffix(configBaseFiles["minio"]), path.Join(cfgPath, cfgVal))
		if err != nil {
			panic(fmt.Errorf("error when reading  minio config: %v", err))
		}
		sources, err = readSourcesGoCSV(configBaseFiles["sources"], path.Join(cfgPath, cfgVal))
		if err != nil {
			panic(fmt.Errorf("error when reading source.csv: %v", err))
		}
	}

	if cfggen {
		err := generateCfg(g1, n1, m1, sources, path.Join(cfgPath, cfgVal), sourcesVal)
		if err != nil {
			fmt.Println(err)
		}
	}
	// validate config
	if cfgval {
		err := valCfg()
		if err != nil {
			fmt.Println(err)
		}
	}

	// load config, on load minio will call webhook in gleaner to run config
	if cfgload {
		err := loadCfg()
		if err != nil {
			fmt.Println(err)
		}
	}

}

func initCfg(cfgpath string, cfgName string, configBaseFiles map[string]string) error {
	fmt.Println("make a config template is there isn't already one")
	var basePath = path.Join(cfgpath, cfgName)
	if _, err := os.Stat(basePath); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(basePath, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
	for _, f := range configBaseFiles {
		var template = path.Join(cfgpath, cfgName, f)
		var config = path.Join(cfgpath, "template", f)
		copy(config, template)
	}

	return nil
}

func generateCfg(gleaner *viper.Viper, nabu *viper.Viper, minioConfig *viper.Viper, sources []Source, cfgPath string, sourcesVal string) error {
	var err error
	//var sm []SourceConfig
	//var minio MinoConfig

	var mi interface{}
	var date string
	currentTime := time.Now()
	date = currentTime.Format("20060102")
	// sources
	// need a check to see if it is an absolute path, so read not needed, and
	fmt.Println("make copy of sources")
	var original = path.Join(cfgPath, sourcesVal)
	var config = path.Join(cfgPath, date+sourcesVal)
	_, err = copy(original, config)
	if err != nil {
		panic(fmt.Errorf("error when copying sources: %v", err))
	}
	// load minio
	mi = minioConfig.Get("minio")
	// no idea why the unmarshall is not working
	// basically means env substitution needs to be handled by us
	//err = minioConfig.UnmarshalKey( "minio",&minio)
	//if err != nil {
	//	panic(fmt.Errorf("error when writing config: %v", err))
	//}

	fmt.Println("Regnerate gleaner")
	gleaner.SetConfigType("yaml")
	var fn = path.Join(cfgPath, date+gleanerBaseName) // copy previous
	err = gleaner.WriteConfigAs(fn)
	if err != nil {
		panic(fmt.Errorf("error when writing config: %v", err))
	}

	gleaner.Set("minio", mi)
	gleaner.Set("sources", sources)
	// hack to get rid of the sourcetype
	//err =  gleaner.UnmarshalKey("sitemaps", &sm)
	//gleaner.Set("sitemaps", sm)
	fn = path.Join(cfgPath, gleanerBaseName)
	err = gleaner.WriteConfigAs(fn)
	if err != nil {
		panic(fmt.Errorf("error when writing config: %v", err))
	}

	fmt.Println("Regnerate nabu")
	nabu.SetConfigType("yaml")
	fn = path.Join(cfgPath, date+nabuBaseName) // copy previous
	err = nabu.WriteConfigAs(fn)
	if err != nil {
		panic(fmt.Errorf("error when writing config: %v", err))
	}
	nabu.Set("minio", mi)
	var prefix []string
	for _, s := range sources {
		if s.Active {
			prefix = append(prefix, s.Name)
		}
	}
	nabu.Set("objects.prefix", prefix)
	var prefixOff []string
	for _, s := range sources {
		if !s.Active {
			prefixOff = append(prefixOff, s.Name)
		}
	}
	nabu.Set("objects.prefixOff", prefixOff)
	//nabu.Set("sitemaps", sources)
	//// hack to get rid of the sourcetype
	//err =  nabu.UnmarshalKey("sitemaps", &sm)
	//nabu.Set("sitemaps", sm)
	fn = path.Join(cfgPath, nabuBaseName)
	err = nabu.WriteConfigAs(fn)
	if err != nil {
		panic(fmt.Errorf("error when writing config: %v", err))
	}
	return err
}

func valCfg() error {
	fmt.Println("validate config")
	return nil
}

func loadCfg() error {
	fmt.Println("load config")
	return nil
}

//func initBucket()   error {
//	// Set up minio and initialize client
//	endpoint := "192.168.2.131:9000"
//	accessKeyID := ""
//	secretAccessKey := ""
//	useSSL := false
//	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
//	if err != nil {
//		log.Fatalln(err)
//	}
//
//	// Make a new bucket called mymusic.
//	bucketName := "onlyatest"
//	location := ""
//
//	err = minioClient.MakeBucket(bucketName, location)
//	if err != nil {
//		// Check to see if we already own this bucket (which happens if you run this twice)
//		exists, err := minioClient.BucketExists(bucketName)
//		if err == nil && exists {
//			log.Printf("We already own %s\n", bucketName)
//		} else {
//			log.Fatalln(err)
//		}
//	} else {
//		log.Printf("Successfully created %s\n", bucketName)
//	}
//
//	err = setNotification(minioClient)
//	if err != nil {
//		log.Fatalln(err)
//	}
//
//	return err
//}

//func setNotification(minioClient *minio.Client) error {
//	queueArn := minio.NewArn("minio", "sqs", "", "1", "webhook")
//	// arn:minio:sqs::1:webhook
//
//	queueConfig := minio.NewNotificationConfig(queueArn)
//	queueConfig.AddEvents(minio.ObjectCreatedAll, minio.ObjectRemovedAll)
//	// queueConfig.AddFilterPrefix("photos/")
//	queueConfig.AddFilterSuffix(".json")
//
//	bucketNotification := minio.BucketNotification{}
//	bucketNotification.AddQueue(queueConfig)
//
//	err := minioClient.SetBucketNotification("onlyatest", bucketNotification)
//	if err != nil {
//		fmt.Println("Unable to set the bucket notification: ", err)
//		return nil
//	}
//
//	return err
//}

func readConfig(filename string, cfgPath string, defaults map[string]interface{}) (*viper.Viper, error) {
	v := viper.New()
	for key, value := range defaults {
		v.SetDefault(key, value)
	}
	v.SetConfigName(filename)
	v.AddConfigPath(cfgPath)
	v.AutomaticEnv()
	err := v.ReadInConfig()
	return v, err
}

func readConfigNoDefault(filename string, cfgPath string) (*viper.Viper, error) {
	v := viper.New()

	v.SetConfigName(filename)
	v.AddConfigPath(cfgPath)
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

func readSources(filename string) ([]Source, error) {
	var sources []Source
	var err error
	r := easycsv.NewReaderFile(filename)
	for r.Read(&sources) {
		fmt.Print(sources)
	}
	if err := r.Done(); err != nil {
		log.Fatalf("Failed to read a CSV file: %v", err)
	}
	return sources, err

}
func readSourcesGoCSV(filename string, cfgPath string) ([]Source, error) {
	var sources []Source
	var err error
	var fn = path.Join(cfgPath, filename)
	f, err := os.Open(fn)
	if err != nil {
		log.Fatal(err)
	}

	// remember to close the file at the end of the program
	defer f.Close()

	if err := gocsv.Unmarshal(f, &sources); err != nil {
		fmt.Println("error:", err)
	}

	for _, u := range sources {
		fmt.Printf("%+v\n", u)
	}
	return sources, err

}

func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func fileNameWithoutExtTrimSuffix(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}
