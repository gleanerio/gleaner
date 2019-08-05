package shapes

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"

	"earthcube.org/Project418/gleaner/internal/common"
	"earthcube.org/Project418/gleaner/internal/millers/millerutils"
	"earthcube.org/Project418/gleaner/pkg/utils"

	"github.com/knakk/rdf"
	minio "github.com/minio/minio-go"
)

// SHACLMillObjects test a concurrent version of calling mock
func SHACLMillObjects(mc *minio.Client, bucketname string, cs utils.Config) {
	entries := common.GetMillObjects(mc, bucketname)
	multiCall(entries, bucketname, mc, cs)
}

func multiCall(e []common.Entry, bucketname string, mc *minio.Client, cs utils.Config) {
	semaphoreChan := make(chan struct{}, 10) // a blocking channel to keep concurrency under control (1 == single thread)
	defer close(semaphoreChan)
	wg := sync.WaitGroup{} // a wait group enables the main process a wait for goroutines to finish

	var gb common.Buffer
	m := common.GetShapeGraphs(mc, "gleaner-shacl") // todo: beware static bucket lists!!!

	for j := range m {
		log.Printf("Checking data graphs against shape graph: %s\n", m[j])
		for k := range e {
			wg.Add(1)
			// log.Printf("Ready JSON-LD package  #%d #%s \n", j, e[k].Urlval)
			go func(j, k int) {
				semaphoreChan <- struct{}{}
				status := shaclTest(e[k].Urlval, e[k].Jld, m[j].Key, m[j].Jld, &gb)

				wg.Done() // tell the wait group that we be done
				log.Printf("#%d #%s wrote %d bytes", j, e[k].Urlval, status)

				<-semaphoreChan
			}(j, k)
		}
	}
	wg.Wait()

	log.Println(gb.Len())

	// TODO   gb is type turtle here..   need to convert to ntriples to store
	nt, err := rdf2rdf(gb.String())
	if err != nil {
		log.Println(err)
	}

	// write to S3
	_, err = millerutils.LoadToMinio(nt, "gleaner-milled", fmt.Sprintf("%s/%s_shacl.nt", cs.Gleaner.RunID, bucketname), mc)
	if err != nil {
		log.Println(err)
	}

}

func rdf2rdf(r string) (string, error) {
	// Decode the existing triples
	var inFormat rdf.Format
	inFormat = rdf.Turtle

	var outFormat rdf.Format
	outFormat = rdf.NTriples

	var s string
	buf := bytes.NewBufferString(s)

	dec := rdf.NewTripleDecoder(strings.NewReader(r), inFormat)
	tr, err := dec.DecodeAll()

	enc := rdf.NewTripleEncoder(buf, outFormat)
	err = enc.EncodeAll(tr)

	enc.Close()

	return buf.String(), err
}

func shaclTest(urlval, dg, sgkey, sg string, gb *common.Buffer) int {
	datagraph, err := millerutils.JSONLDToTTL(dg, urlval)
	if err != nil {
		log.Printf("Error in the jsonld write... %v\n", err)
		log.Printf("Nothing to do..   going home")
		return 0
	}

	url := "http://localhost:7000"
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("dataref", urlval)
	writer.WriteField("shaperef", sgkey)

	part, err := writer.CreateFormFile("datag", "datagraph")
	if err != nil {
		log.Println(err)
	}
	_, err = io.Copy(part, strings.NewReader(datagraph))
	if err != nil {
		log.Println(err)
	}

	part, err = writer.CreateFormFile("shapeg", "shapegraph")
	if err != nil {
		log.Println(err)
	}
	_, err = io.Copy(part, strings.NewReader(sg))
	if err != nil {
		log.Println(err)
	}

	err = writer.Close()
	if err != nil {
		log.Println(err)
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		log.Println(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", "EarthCube_DataBot/1.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	// write result to buffer
	len, err := gb.Write(b)
	if err != nil {
		log.Printf("error in the buffer write... %v\n", err)
	}

	return len //  we will return the bytes count we write...
}
