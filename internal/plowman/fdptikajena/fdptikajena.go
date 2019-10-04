package fdptikajena

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"earthcube.org/Project418/gleaner/internal/common"
	"earthcube.org/Project418/gleaner/internal/millers/millerutils"
	"earthcube.org/Project418/gleaner/pkg/utils"

	// "github.com/bbalet/stopwords"
	"github.com/deiu/rdf2go"
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
func TikaObjects(mc *minio.Client, bucketname string, cs utils.Config) {
	graphname := fmt.Sprintf("%s", bucketname)
	entries := common.GetMillObjects(mc, bucketname)
	multiCall(mc, entries, graphname, cs)
}

func multiCall(mc *minio.Client, e []common.Entry, graphname string, cs utils.Config) {
	// Set up the the semaphore and conccurancey
	semaphoreChan := make(chan struct{}, 4) // a blocking channel to keep concurrency under control
	defer close(semaphoreChan)
	wg := sync.WaitGroup{} // a wait group enables the main process a wait for goroutines to finish

	var gb common.Buffer // use later to allow a mutex locked []byte

	for k := range e {
		wg.Add(1)
		log.Printf("Setting up Tika Package call #%d in a goroutine\n", k)
		go func(k int) {
			semaphoreChan <- struct{}{}

			status := tikaIndex(e[k].Bucketname, e[k].Key, e[k].Urlval, e[k].Jld, graphname, &gb)

			wg.Done() // tell the wait group that we be done
			log.Printf("#%d done with %s for package LP %s", k, status, e[k].Urlval)
			log.Println(gb.Len())

			<-semaphoreChan
		}(k)
	}
	wg.Wait()

	log.Println(gb.Len())

	// write to S3
	fl, err := millerutils.LoadToMinio(gb.String(), "gleaner-milled", fmt.Sprintf("%s/%s_fdpjena.n3", cs.Gleaner.RunID, graphname), mc)
	// deprecated write to file
	// fl, err := millerutils.WriteRDF(gb.String(), fmt.Sprintf("%s_fdpjena", graphname))
	if err != nil {
		log.Println("RDF file could not be written")
	} else {
		log.Printf("RDF file for Jena/Fuseki and Lucene written len:%d\n", fl)
	}
}

func tikaIndex(bucketname, key, urlval, jsonld, graphname string, gb *common.Buffer) string {
	_, m := getBytes(urlval, "datapackage.json")

	// Set up the the semaphore and conccurancey
	semaphoreChan := make(chan struct{}, 6) // a blocking channel to keep concurrency under control
	defer close(semaphoreChan)
	wg := sync.WaitGroup{} // a wait group enables the main process a wait for goroutines to finish

	var lb common.Buffer // use later to allow a mutex locked []byte

	ms := parsePackage(string(m))
	for k := range ms.Resources {
		wg.Add(1)
		log.Printf("Adding tika index go routine #%d for resource %s\n", k, ms.Resources[k].Name)
		go func(k int) {
			semaphoreChan <- struct{}{}

			sw := callTika(&lb, urlval, ms.Resources[k].Path)

			wg.Done() // tell the wait group that we be done
			log.Printf("#%d status %s for %s / %s", k, sw, ms.Resources[k].Name, urlval)
			<-semaphoreChan
		}(k)
	}
	wg.Wait()

	// TODO copy lb (local buffer) to gb (global buffer)
	// return OK or len gb?
	log.Println("In tikaIndex at the gbWrite line")
	_, err := gb.Write([]byte(lb.String()))
	if err != nil {
		log.Printf("error in the lb buffer write... %v\n", err)
	}

	return "ok"
}

func callTika(lb *common.Buffer, urlval, path string) string {
	url := "http://localhost:9998/tika" // default  is 9998  (I use 80 with ha proxy)

	// convert _ to s (status and use !=200 to return "false")
	_, b := getBytes(urlval, path)

	req, err := http.NewRequest("PUT", url, bytes.NewReader(b))
	if err != nil {
		log.Printf("http.NewRequest %v\n", err)
		return "error"
	}
	req.Header.Set("Accept", "text/plain")
	req.Header.Set("User-Agent", "EarthCube_DataBot/1.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("client.Do %v\n", err)
		return "error"
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ioutil.ReadAll  %v\n", err)
		return "error"
	}

	// TODO..   skip stop word creation and see how big it makes things...
	// sw := stopwords.CleanString(string(body), "en", true) // remove stop words..   no reason for them in the search

	re := regexp.MustCompile(`\r?\n?\t`)
	sw := re.ReplaceAllString(string(body), " ")
	sw = strings.Replace(sw, "\n", " ", -1)
	//input = re.ReplaceAllString(input, " ")

	// build a graph here...
	g := rdf2go.NewGraph("")
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(fmt.Sprintf("%s/%s", urlval, path)),
		rdf2go.NewResource("http://www.w3.org/2000/01/rdf-schema#comment"),
		rdf2go.NewLiteral(sw))) //  If we turn stop words back on, use sw here

	// TODO
	// Put in refernce triples
	//g.Add(rdf2go.NewTriple(rdf2go.NewResource(fmt.Sprintf("%s/%s", urlval, path)),
	//	rdf2go.NewResource("http://www.w3.org/2000/01/rdf-schema#comment"),
	//	rdf2go.NewLiteral(sw))) //  If we turn stop words back on, use sw here

	_, err = lb.Write([]byte(g.String()))
	if err != nil {
		log.Printf("error in the lb buffer write... %v\n", err)
		return "error"
	}

	return "ok"
}

func getBytes(url, key string) (int, []byte) {
	resurl := fmt.Sprintf("%s/%s", url, key)
	client := resty.New()
	resp, err := client.R().Get(resurl)
	if err != nil {
		log.Printf("getBytes %v\n", err)
	}
	return resp.StatusCode(), resp.Body()
}

func parsePackage(j string) Manifest {
	m := Manifest{}
	json.Unmarshal([]byte(j), &m)
	return m
}
