package ner

import (
	//	"bytes"

	"fmt"
	"log"
	"sync"

	//	"log"

	"earthcube.org/Project418/gleaner/internal/common"
	"github.com/buger/jsonparser"
	minio "github.com/minio/minio-go"
	prose "gopkg.in/jdkato/prose.v2"
)

// MockObjects test a concurrent version of calling mock
func NERObjects(mc *minio.Client, bucketname string) {
	entries := common.GetMillObjects(mc, bucketname)
	multiCall(entries)
}

func multiCall(e []common.Entry) {
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
	fmt.Printf("%s:  %s %s   %s =? %s \n", bucketname, key, urlval, sha1val, doner(jsonld))
	return "ok"
}

func doner(jsonld string) string {
	//  for NER lets just pull the description
	dl, err := jsonparser.GetString([]byte(jsonld), "description")
	if err != nil {
		log.Println(err)
		return ""
	}

	bss := ""
	doc, _ := prose.NewDocument(dl)
	for _, ent := range doc.Entities() {
		bss = fmt.Sprintf("%s %s", ent.Text, ent.Label)
	}

	return bss
}
