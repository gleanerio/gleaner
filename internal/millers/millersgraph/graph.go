package millersgraph

import (
	"fmt"
	"log"
	"sync"

	"earthcube.org/Project418/gleaner/internal/millers/millerutils"
	"earthcube.org/Project418/gleaner/internal/utils"

	minio "github.com/minio/minio-go"
)

// GraphMillObjects test a concurrent version of calling mock
func GraphMillObjects(mc *minio.Client, bucketname string) {
	entries := utils.GetMillObjects(mc, bucketname)
	multiCall(entries, bucketname, mc)
}

func multiCall(e []utils.Entry, bucketname string, mc *minio.Client) {
	// Set up the the semaphore and conccurancey
	semaphoreChan := make(chan struct{}, 20) // a blocking channel to keep concurrency under control
	defer close(semaphoreChan)
	wg := sync.WaitGroup{} // a wait group enables the main process a wait for goroutines to finish

	var gb utils.Buffer

	for k := range e {
		wg.Add(1)
		log.Printf("About to run #%d in a goroutine\n", k)
		go func(k int) {
			semaphoreChan <- struct{}{}
			status := millerutils.Jsl2graph(e[k].Bucketname, e[k].Key, e[k].Urlval, e[k].Sha1val, e[k].Jld, &gb)

			wg.Done() // tell the wait group that we be done
			log.Printf("#%d wrote %d bytes", k, status)
			<-semaphoreChan
		}(k)
	}
	wg.Wait()

	log.Println(gb.Len())

	// write to S3
	_, err := millerutils.LoadToMinio(gb.String(), "gleaner", fmt.Sprintf("%s.n3", bucketname), mc)

	// write to file
	fl, err := millerutils.WriteRDF(gb.String(), bucketname)
	if err != nil {
		log.Println("RDF file could not be written")
	} else {
		log.Printf("RDF file written len:%d\n", fl)
	}

}
