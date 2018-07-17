package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"earthcube.org/Project418/gleaner/millers/millerbleve"
	"earthcube.org/Project418/gleaner/millers/millerfdpgraph"
	"earthcube.org/Project418/gleaner/millers/millerfdptika"
	"earthcube.org/Project418/gleaner/millers/millerprov"
	"earthcube.org/Project418/gleaner/millers/millersgraph"
	"earthcube.org/Project418/gleaner/millers/millersmock"
	"earthcube.org/Project418/gleaner/millers/millerspatial"
	"earthcube.org/Project418/gleaner/millers/millertika"
	"earthcube.org/Project418/gleaner/millers/millerutils"

	"earthcube.org/Project418/gleaner/summoner/acquire"
	"earthcube.org/Project418/gleaner/utils"
	minio "github.com/minio/minio-go"
)

func main() {
	// TODO
	// One unified configuration file
	// time stamped minio buckets   DDMMYYYY  make bucket: 22052018Products  --> then just the expected names number that
	// save triples to a minio bucket (enable remote)
	// copy tile38 AOG to minio on completion
	// _move_  (copy?) bleve indexes to minio on completion

	// flags
	millPtr := flag.Bool("mill", false, "a bool to activate milling")
	summonPtr := flag.Bool("summon", false, "a bool to request summoning ")
	cfgFileLoc := flag.String("config", "config.json", "JSON configure file")
	flag.Parse()

	// load config file
	cs := utils.LoadConfiguration(cfgFileLoc)

	// TODO check for output directory and make it if it doesn't exist
	path := "./output" // put in config file????
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, os.ModePerm)
	}

	// TODO set up an output directory for each run in output (date)
	// move RDF writer func to util directory
	// pass all functions this value to write to
	t := time.Now()
	year, month, day := t.Date()
	hour, min, sec := t.Clock()

	rundir := fmt.Sprintf("%s/%d%d%d_%d%d%d", path, year, month, day, hour, min, sec)
	if _, err := os.Stat(rundir); os.IsNotExist(err) {
		os.Mkdir(rundir, os.ModePerm)
	}

	// Set up our log file for runs...
	logfile := fmt.Sprintf("%s/logfile.txt", rundir)
	f, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	if *summonPtr {
		summoner(cs)
	}

	if *millPtr {
		millers(cs, rundir)
	}
}

func millers(cs utils.Config, rundir string) {

	millerutils.RunDir = rundir // set output dir for graph, fdpgraph and prov
	// set output for bleve and fdptika and tika

	st := time.Now()
	log.Printf("Miller start time: %s \n", st) // Log the time at start for the record

	mc := miniConnection(cs) // minio connection

	buckets, err := listBuckets(mc)
	if err != nil {
		log.Println(err)
	}

	fmt.Println("Bucket list...")
	for _, bucket := range buckets {
		log.Println(bucket.Name) // for no real reason.. :)
	}

	// Make an array "as" of active buckets to process...
	as := []string{}
	for i := range cs.Sources {
		if cs.Sources[i].Active == true {
			as = append(as, cs.Sources[i].Name)
		}
	}

	// Mock is just a template miller..  prints resource entries only...
	if cs.Millers.Mock {
		for d := range as {
			millersmock.MockObjects(mc, as[d])
		}
	}

	if cs.Millers.Graph {
		for d := range as {
			millersgraph.GraphMillObjects(mc, as[d])
		}
	}

	if cs.Millers.Spatial {
		for d := range as {
			millerspatial.ProcessBucketObjects(mc, as[d])
		}
		// TODO add in saving the AOF file to the output directory
	}

	if cs.Millers.Organic {
		for d := range as {
			millerbleve.GetObjects(mc, as[d])
		}
	}

	if cs.Millers.Prov {
		for d := range as {
			millerprov.MockObjects(mc, as[d])
		}
	}

	if cs.Millers.Tika {
		for d := range as {
			millertika.TikaObjects(mc, as[d])
		}
	}

	if cs.Millers.FDPTika {
		for d := range as {
			millerfdptika.TikaObjects(mc, as[d])
		}
	}

	if cs.Millers.FDPGraph {
		for d := range as {
			millerfdpgraph.TikaObjects(mc, as[d])
		}
	}

	et := time.Now()
	diff := et.Sub(st)
	log.Printf("Miller end time: %s \n", et)
	log.Printf("Miller run time: %f \n", diff.Minutes())
}

func summoner(cs utils.Config) {
	log.Printf("Summoner start time: %s \n", time.Now()) // Log start time

	domains, headlessdomains, err := acquire.DomainListJSON(cs)
	if err != nil {
		log.Printf("Error reading list of domains %v\n", err)
	}

	log.Printf("Domains: %s \n", domains)
	log.Printf("Headless domains: %s \n", headlessdomains)

	// TODO  the following two functions could be done concurrently
	ru := acquire.ResourceURLsJSON(domains) // map by domain name and []string of landing page URLs
	if len(ru) > 0 {
		acquire.ResRetrieve(ru, cs)
	}

	hru := acquire.ResourceURLsJSON(headlessdomains) // map by domain name and []string of landing page URLs
	if len(hru) > 0 {
		acquire.Headless(hru, cs)
	}

	log.Printf("Summoner end time: %s \n", time.Now()) // Log end time
}

// Set up minio and initialize client
func miniConnection(cs utils.Config) *minio.Client {
	endpoint := cs.Minio.Endpoint
	accessKeyID := cs.Minio.AccessKeyID
	secretAccessKey := cs.Minio.SecretAccessKey
	useSSL := false
	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		log.Fatalln(err)
	}
	return minioClient
}

func listBuckets(mc *minio.Client) ([]minio.BucketInfo, error) {
	buckets, err := mc.ListBuckets()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return buckets, err
}
