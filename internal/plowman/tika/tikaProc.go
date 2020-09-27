package tika

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	//	"log"

	"github.com/bbalet/stopwords"
	"github.com/blevesearch/bleve"
	"github.com/buger/jsonparser"
	"github.com/earthcubearchitecture-project418/gleaner/internal/common"
	"github.com/earthcubearchitecture-project418/gleaner/internal/millers/millerutils"
	minio "github.com/minio/minio-go"
)

// MockObjects test a concurrent version of calling mock
func TikaObjects(mc *minio.Client, bucketname string) {
	// indexname := fmt.Sprintf("./output/bleve/%s_data", bucketname)
	// initBleve(indexname)
	// entries := utils.GetMillObjects(mc, bucketname)
	// multiCall(entries, indexname)

	indexname := fmt.Sprintf("%s_data", bucketname)
	fp := millerutils.NewinitBleve(indexname) //  initBleve(indexname)
	entries := common.GetMillObjects(mc, bucketname)
	multiCall(entries, fp)
}

// Initialize the text index  // this function needs some attention (of course they all do)
// func initBleve(filename string) {
// 	mapping := bleve.NewIndexMapping()
// 	index, berr := bleve.New(filename, mapping)
// 	if berr != nil {
// 		log.Printf("Bleve error making index %v \n", berr)
// 	}
// 	index.Close()
// }

func multiCall(e []common.Entry, indexname string) {
	// TODO..   open the bleve index here once and pass by reference to text
	index, berr := bleve.Open(indexname)
	if berr != nil {
		// should panic here?..  no index..  no reason to keep living  :(
		log.Printf("Bleve error making index %v \n", berr)
	}

	// Set up the the semaphore and conccurancey
	semaphoreChan := make(chan struct{}, 10) // a blocking channel to keep concurrency under control
	defer close(semaphoreChan)
	wg := sync.WaitGroup{} // a wait group enables the main process a wait for goroutines to finish

	for k := range e {
		wg.Add(1)
		log.Printf("About to run #%d in a goroutine\n", k)
		go func(k int) {
			semaphoreChan <- struct{}{}

			status := tikaIndex(e[k].Bucketname, e[k].Key, e[k].Urlval, e[k].Jld, index)

			wg.Done() // tell the wait group that we be done
			log.Printf("#%d done with %s with %s", k, status, e[k].Urlval)
			<-semaphoreChan
		}(k)
	}
	wg.Wait()

	index.Close()
}

// Mock is a simple function to use as a stub for talking about millers
func tikaIndex(bucketname, key, urlval, jsonld string, index bleve.Index) string {
	// Pull The file download URLs from the jsonld
	// dl, err := jsonparser.GetString([]byte(jsonld), "distribution", "contentUrl")
	dl, err := jsonparser.GetString([]byte(jsonld), "url")
	if err != nil {
		log.Println(err)
		return "bad"
	}

	// BCO-DMO filter for urls with only
	// if strings.Contains(dl, "/dataset/") != true {
	// 	log.Println("Skipping non-data URL")
	// 	return "skipped"
	// }

	// TODO
	// get the mimetype too..   then only process the file is they map a type if we want
	// or pass the mimetype along to the process.
	// mt, err := jsonparser.GetString([]byte(jsonld), "distribution", "fileType??")
	// if err != nil {
	// 	log.Println(err)
	// 	return "bad"
	// }

	// TODO
	// Given the URL..   get the datapackage file
	// convert to a struct..
	// loop on entries....
	// index them one by one with URL to download...

	rd, err := http.Get(dl)
	if err != nil {
		log.Println(err)
		return "bad"
	}
	defer rd.Body.Close()

	url := "http://localhost:9998/tika"
	r, _ := ioutil.ReadAll(rd.Body)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(r))
	req.Header.Set("Accept", "text/plain")
	req.Header.Set("User-Agent", "EarthCube_DataBot/1.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	// fmt.Println("Tika Response Status:", resp.Status)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	sw := stopwords.CleanString(string(body), "en", true)

	// fmt.Println(urlval)
	// fmt.Println(dl)
	// fmt.Println(sw)

	// index some data
	berr := index.Index(urlval, sw)
	log.Printf("Bleve Indexed item with ID %s\n", urlval)
	if berr != nil {
		log.Printf("Bleve error indexing %v \n", berr)
	}

	// load the cleanstring into KV store for later belve indexing in stage 2.

	// then open and save to a bleve index like in the beleve indexer...
	// NOTE like in the beleve indexer this then has to be single threaded or save
	// results to a KV store to later sequentially index via belve.

	return "ok"
}
