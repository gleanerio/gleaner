package millerfdpgraph

import (
	"fmt"
	"log"
	"sync"

	"earthcube.org/Project418/gleaner/millers/millerutils"
	"earthcube.org/Project418/gleaner/utils"
	minio "github.com/minio/minio-go"
)

// Manifest is the struct for the manifest from the data package
// do not need the full datapackage.json, just the file manifest
type Manifest struct {
	Profile   string `json:"profile"`
	Resources []struct {
		Encoding string `json:"encoding"`
		Name     string `json:"name"`
		Path     string `json:"path"`
		Profile  string `json:"profile"`
	} `json:"resources"`
}

// TikaObjects test a concurrent version of calling mock
func TikaObjects(mc *minio.Client, bucketname string) {
	entries := utils.GetMillObjects(mc, bucketname)
	multiCall(entries, bucketname)
}

func multiCall(e []utils.Entry, bucketname string) {

	// Set up the the semaphore and conccurancey
	semaphoreChan := make(chan struct{}, 20) // a blocking channel to keep concurrency under control
	defer close(semaphoreChan)
	wg := sync.WaitGroup{} // a wait group enables the main process a wait for goroutines to finish

	var gb utils.Buffer // use later to allow a mutex locked []byte

	for k := range e {
		wg.Add(1)
		log.Printf("About to run #%d in a goroutine\n", k)
		go func(k int) {
			semaphoreChan <- struct{}{}

			status := millerutils.Jsl2graph(e[k].Bucketname, e[k].Key, e[k].Urlval, e[k].Sha1val, e[k].Jld, &gb)

			wg.Done() // tell the wait group that we be done
			log.Printf("#%d done with %s with %s", k, status, e[k].Urlval)
			<-semaphoreChan
		}(k)
	}
	wg.Wait()

	log.Println(gb.Len())
	fl, err := millerutils.WriteRDF(gb.String(), fmt.Sprintf("%s_fdp", bucketname))
	if err != nil {
		log.Println("RDF file could not be written")
	} else {
		log.Printf("RDF file written len:%d\n", fl)
	}
}
