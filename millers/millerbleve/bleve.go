package millerbleve

import (
	"fmt"
	"log"
	"sync"

	"earthcube.org/Project418/gleaner/millers/millerutils"
	"earthcube.org/Project418/gleaner/utils"
	"github.com/blevesearch/bleve"
	minio "github.com/minio/minio-go"
)

// GetObjects test a concurrent version of calling mock
func GetObjects(mc *minio.Client, bucketname string) {
	indexname := fmt.Sprintf("%s", bucketname)
	fp := millerutils.NewinitBleve(indexname) //  initBleve(indexname)
	entries := utils.GetMillObjects(mc, bucketname)
	multiCall(entries, fp)

}

// // Initialize the text index  // this function needs some attention (of course they all do)
// func initBleve(filename string) {
// 	mapping := bleve.NewIndexMapping()
// 	index, berr := bleve.New(filename, mapping)
// 	if berr != nil {
// 		log.Printf("Bleve error making index %v \n", berr)
// 	}
// 	index.Close()
// }

func multiCall(e []utils.Entry, indexname string) {
	// TODO..   open the bleve index here once and pass by reference to text
	index, berr := bleve.Open(indexname)
	if berr != nil {
		// should panic here?..  no index..  no reason to keep living  :(
		log.Printf("Bleve error making index %v \n", berr)
	}

	// Set up the the semaphore and conccurancey
	semaphoreChan := make(chan struct{}, 1) //For direct write like this must be SINGLE THREADED!!!!!!
	defer close(semaphoreChan)
	wg := sync.WaitGroup{} // a wait group enables the main process a wait for goroutines to finish

	for k := range e {
		wg.Add(1)
		log.Printf("About to run #%d in a goroutine\n", k)
		go func(k int) {
			semaphoreChan <- struct{}{}

			status := textIndexer(e[k].Urlval, e[k].Jld, index)

			wg.Done() // tell the wait group that we be done
			log.Printf("#%d done with %s", k, status)
			<-semaphoreChan
		}(k)
	}
	wg.Wait()

	index.Close()
}

// index some jsonld with an ID
func textIndexer(ID string, jsonld string, index bleve.Index) string {
	berr := index.Index(ID, jsonld)
	log.Printf("Blevel Indexed item with ID %s\n", ID)
	if berr != nil {
		log.Printf("Bleve error indexing %v \n", berr)
	}

	return "done"
}
