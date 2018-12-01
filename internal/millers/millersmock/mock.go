package millersmock

import (
	//	"bytes"

	"crypto/sha1"
	"fmt"
	"log"
	"sync"

	//	"log"

	"earthcube.org/Project418/gleaner/internal/utils"
	minio "github.com/minio/minio-go"
)

// MockObjects test a concurrent version of calling mock
func MockObjects(mc *minio.Client, bucketname string) {
	entries := utils.GetMillObjects(mc, bucketname)
	multiCall(entries)
}

func multiCall(e []utils.Entry) {
	// Set up the the semaphore and conccurancey
	semaphoreChan := make(chan struct{}, 20) // a blocking channel to keep concurrency under control
	defer close(semaphoreChan)
	wg := sync.WaitGroup{} // a wait group enables the main process a wait for goroutines to finish

	for k := range e {
		wg.Add(1)
		fmt.Printf("About to run #%d in a goroutine\n", k)
		go func(k int) {
			semaphoreChan <- struct{}{}
			status := simplePrint(e[k].Bucketname, e[k].Key, e[k].Urlval, e[k].Sha1val, e[k].Jld)

			wg.Done() // tell the wait group that we be done
			log.Printf("#%d done with %s", k, status)
			<-semaphoreChan
		}(k)
	}
	wg.Wait()
}

// Mock is a simple function to use as a stub for talking about millers
func simplePrint(bucketname, key, urlval, sha1val, jsonld string) string {
	fmt.Printf("%s:  %s %s   %s =? %s \n", bucketname, key, urlval, sha1val, getsha(jsonld))
	return "ok"
}

func getsha(jsonld string) string {
	h := sha1.New()
	h.Write([]byte(jsonld))
	bs := h.Sum(nil)
	bss := fmt.Sprintf("%x", bs)
	return bss
}
