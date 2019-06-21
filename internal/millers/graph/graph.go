package graph

import (
	"fmt"
	"log"
	"sync"

	"earthcube.org/Project418/gleaner/internal/common"
	"earthcube.org/Project418/gleaner/internal/millers/millerutils"
	"earthcube.org/Project418/gleaner/pkg/utils"

	minio "github.com/minio/minio-go"
)

// GraphMillObjects test a concurrent version of calling mock
func GraphMillObjects(mc *minio.Client, bucketname string, cs utils.Config) {
	entries := common.GetMillObjects(mc, bucketname)
	multiCall(entries, bucketname, mc, cs)
}

func multiCall(e []common.Entry, bucketname string, mc *minio.Client, cs utils.Config) {
	// Set up the the semaphore and conccurancey
	semaphoreChan := make(chan struct{}, 20) // a blocking channel to keep concurrency under control
	defer close(semaphoreChan)
	wg := sync.WaitGroup{} // a wait group enables the main process a wait for goroutines to finish

	var gb common.Buffer

	for k := range e {
		wg.Add(1)
		log.Printf("About to run #%d in a goroutine\n", k)
		go func(k int) {
			semaphoreChan <- struct{}{}
			status := millerutils.Jsl2graph(e[k].Bucketname, e[k].Key, e[k].Urlval, e[k].Sha1val, e[k].Jld, &gb)

			wg.Done() // tell the wait group that we be done!
			log.Printf("#%d wrote %d bytes", k, status)
			<-semaphoreChan
		}(k)
	}
	wg.Wait()

	log.Println(gb.Len())

	// write to S3
	// fl, err := millerutils.LoadToMinio(gb.String(), fmt.Sprintf("gleaner-milled/%s", cs.Gleaner.RunID), fmt.Sprintf("%s.n3", bucketname), mc)
	fl, err := millerutils.LoadToMinio(gb.String(), "gleaner-milled", fmt.Sprintf("%s/%s.n3", cs.Gleaner.RunID, bucketname), mc)

	// deprecated write to file
	// fl, err := millerutils.WriteRDF(gb.String(), bucketname)
	if err != nil {
		log.Println("RDF file could not be written")
	} else {
		log.Printf("RDF file written len:%d\n", fl)
	}

}
