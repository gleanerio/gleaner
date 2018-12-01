package millerfdptika

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	//	"log"

	"earthcube.org/Project418/gleaner/internal/millers/millerutils"
	"earthcube.org/Project418/gleaner/internal/utils"
	"github.com/bbalet/stopwords"
	"github.com/blevesearch/bleve"
	"github.com/go-resty/resty"
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
	// indexname := fmt.Sprintf("./output/bleve/%s_packages", bucketname)
	// initBleve(indexname)
	indexname := fmt.Sprintf("%s_packages", bucketname)
	fp := millerutils.NewinitBleve(indexname) //  initBleve(indexname)
	entries := utils.GetMillObjects(mc, bucketname)
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

func multiCall(e []utils.Entry, indexname string) {
	// TODO..   open the bleve index here once and pass by reference to text
	index, berr := bleve.Open(indexname)
	if berr != nil {
		// should panic here?..  no index..  no reason to keep living  :(
		log.Printf("Bleve error making index %v \n", berr)
	}

	// Set up the the semaphore and conccurancey
	semaphoreChan := make(chan struct{}, 1) // a blocking channel to keep concurrency under control
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

func tikaIndex(bucketname, key, urlval, jsonld string, index bleve.Index) string {
	_, m := getBytes(urlval, "datapackage.json")

	ms := parsePackage(string(m))
	for _, v := range ms.Resources {
		// fmt.Println(v.Path)
		// fmt.Println(v.Name)
		s, b := getBytes(urlval, v.Path)

		if s == 200 {
			url := "http://localhost:9998/tika"

			req, err := http.NewRequest("PUT", url, bytes.NewReader(b))
			req.Header.Set("Accept", "text/plain")
			req.Header.Set("User-Agent", "EarthCube_DataBot/1.0")

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
			sw := stopwords.CleanString(string(body), "en", true) // remove stop words..   no reason for them in the search

			// index some data
			resurl := fmt.Sprintf("%s/%s", urlval, v.Path)

			berr := index.Index(resurl, sw)
			log.Printf("Bleve Indexed item with ID %s\n", resurl)
			if berr != nil {
				log.Printf("Bleve error indexing %v \n", berr)
			}
		}
	}
	return "ok"
}

func getBytes(url, key string) (int, []byte) {
	resurl := fmt.Sprintf("%s/%s", url, key)
	resp, err := resty.R().Get(resurl)
	if err != nil {
		log.Println(err)
	}
	return resp.StatusCode(), resp.Body()
}

func parsePackage(j string) Manifest {
	m := Manifest{}
	json.Unmarshal([]byte(j), &m)
	return m
}
