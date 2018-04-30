package millertika

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	//	"log"

	"earthcube.org/Project418/gleaner/millers/utils"
	"github.com/bbalet/stopwords"
	"github.com/buger/jsonparser"
	minio "github.com/minio/minio-go"
)

// MockObjects test a concurrent version of calling mock
func TikaObjects(mc *minio.Client, bucketname string) {
	entries := utils.GetMillObjects(mc, bucketname)
	multiCall(entries)
}

func multiCall(e []utils.Entry) {
	// Set up the the semaphore and conccurancey
	semaphoreChan := make(chan struct{}, 1) // a blocking channel to keep concurrency under control
	defer close(semaphoreChan)
	wg := sync.WaitGroup{} // a wait group enables the main process a wait for goroutines to finish

	for k := range e {
		wg.Add(1)
		log.Printf("About to run #%d in a goroutine\n", k)
		go func(k int) {
			semaphoreChan <- struct{}{}
			status := simplePrint(e[k].Bucketname, e[k].Key, e[k].Urlval, e[k].Sha1val, e[k].Jld)

			wg.Done() // tell the wait group that we be done
			log.Printf("#%d done with %s with %s", k, status, e[k].Urlval)
			<-semaphoreChan
		}(k)
	}
	wg.Wait()
}

// Mock is a simple function to use as a stub for talking about millers
func simplePrint(bucketname, key, urlval, sha1val, jsonld string) string {
	//	fmt.Printf("%s:  %s %s   %s =? %s \n", bucketname, key, urlval, sha1val, getsha(jsonld))

	// Pull The file download URLs from the jsonld
	dl, _ := jsonparser.GetString([]byte(jsonld), "distribution", "contentUrl")

	rd, err := http.Get(dl)
	if err != nil {
		log.Println("Can not get the file to download")
		log.Println(err)
		return "bad"
	}
	defer rd.Body.Close()

	url := "http://localhost:9998/tika"
	r, _ := ioutil.ReadAll(rd.Body)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(r))
	req.Header.Set("Accept", "text/plain")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	// fmt.Println("Tika Response Status:", resp.Status)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	sw := stopwords.CleanString(string(body), "en", true)

	fmt.Println(urlval)
	fmt.Println(dl)
	fmt.Println(sw)

	// load the cleanstring into KV store for later belve indexing in stage 2.

	// then open and save to a bleve index like in the beleve indexer...
	// NOTE like in the beleve indexer this then has to be single threaded or save
	// results to a KV store to later sequentially index via belve.

	return "ok"
}

func getsha(jsonld string) string {
	h := sha1.New()
	h.Write([]byte(jsonld))
	bs := h.Sum(nil)
	bss := fmt.Sprintf("%x", bs)
	return bss
}
