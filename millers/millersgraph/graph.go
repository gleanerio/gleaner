package millersgraph

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"earthcube.org/Project418/gleaner/millers/utils"
	minio "github.com/minio/minio-go"
	"github.com/piprate/json-gold/ld"
	"github.com/rs/xid"
)

// GraphMillObjects test a concurrent version of calling mock
func GraphMillObjects(mc *minio.Client, bucketname string) {
	entries := utils.GetMillObjects(mc, bucketname)
	multiCall(entries, bucketname)
}

func multiCall(e []utils.Entry, bucketname string) {
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
			status := jsl2graph(e[k].Bucketname, e[k].Key, e[k].Urlval, e[k].Sha1val, e[k].Jld, &gb)

			wg.Done() // tell the wait group that we be done
			log.Printf("#%d wrote %d bytes", k, status)
			<-semaphoreChan
		}(k)
	}
	wg.Wait()

	log.Println(gb.Len())
	fl, err := writeRDF(gb.String(), bucketname)
	if err != nil {
		log.Println("RDF file could not be written")
	} else {
		log.Printf("RDF file written len:%d\n", fl)
	}

}

func writeRDF(rdf string, bucketname string) (int, error) {
	// for now just append to a file..   later I will send to a triple store
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(fmt.Sprintf("./output/graphs/%s.nq", bucketname), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	fl, err := f.Write([]byte(rdf))
	if err != nil {
		log.Fatal(err)
	}

	if err := f.Close(); err != nil {
		log.Fatal(err)
	}

	return fl, err // always nil,  we will never get here with FATAL..   leave for test..  but later remove to log only
}

// Mock is a simple function to use as a stub for talking about millers
func jsl2graph(bucketname, key, urlval, sha1val, jsonld string, gb *utils.Buffer) int {
	log.Printf("%s:  %s %s   %s =? %s \n", bucketname, key, urlval, sha1val, "foo")

	nq, _ := jsonLDToNQ(jsonld) // TODO replace with NQ from isValid function..  saving time..
	rdf := globalUniqueBNodes(nq)

	lpt := lptriples(rdf, urlval)

	nt := fmt.Sprintf("\n" + rdf + "\n" + lpt)

	len, err := gb.Write([]byte(nt))
	if err != nil {
		log.Printf("error in the buffer write... %v\n", err)
	}

	return len //  we will return the bytes count we write...
}

func lptriples(t, urlval string) string {
	scanner := bufio.NewScanner(strings.NewReader(t))

	us := []string{}
	for scanner.Scan() {
		split := strings.Split(scanner.Text(), " ")
		us = appendIfMissing(us, split[0])
	}

	nt := []string{}
	for i := range us {
		t := fmt.Sprintf("%s <http://www.w3.org/2000/01/rdf-schema#seeAlso> <%s> .", us[i], urlval)
		nt = append(nt, t)
	}

	lpt := strings.Join(nt, "\n")

	return lpt
}

func appendIfMissing(slice []string, i string) []string {
	for _, ele := range slice {
		if ele == i {
			return slice
		}
	}
	return append(slice, i)
}

func jsonLDToNQ(jsonld string) (string, error) {
	proc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")
	options.Format = "application/nquads"

	var myInterface interface{}
	err := json.Unmarshal([]byte(jsonld), &myInterface)
	if err != nil {
		log.Println("Error when transforming JSON-LD document to interface:", err)
		return "", err
	}

	nq, err := proc.ToRDF(myInterface, options) // returns triples but toss them, we just want to see if this processes with no err
	if err != nil {
		log.Println("Error when transforming JSON-LD document to RDF:", err)
		return "", err
	}

	return nq.(string), err
}

func globalUniqueBNodes(nq string) string {

	scanner := bufio.NewScanner(strings.NewReader(nq))
	// make a map here to hold our old to new map
	m := make(map[string]string)

	for scanner.Scan() {
		//fmt.Println(scanner.Text())
		// parse the line
		split := strings.Split(scanner.Text(), " ")
		sold := split[0]
		oold := split[2]

		if strings.HasPrefix(sold, "_:") { // we are a blank node
			// check map to see if we have this in our value already
			if _, ok := m[sold]; ok {
				// fmt.Printf("We had %s, already\n", sold)
			} else {
				guid := xid.New()
				snew := fmt.Sprintf("_:b%s", guid.String())
				m[sold] = snew
			}
		}

		// scan the object nodes too.. though we should find nothing here.. the above wouldn't
		// eventually find
		if strings.HasPrefix(oold, "_:") { // we are a blank node
			// check map to see if we have this in our value already
			if _, ok := m[oold]; ok {
				// fmt.Printf("We had %s, already\n", oold)
			} else {
				guid := xid.New()
				onew := fmt.Sprintf("_:b%s", guid.String())
				m[oold] = onew
			}
		}
		// triple := tripleBuilder(split[0], split[1], split[3])
		// fmt.Println(triple)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	//fmt.Println(m)

	filebytes := []byte(nq)

	for k, v := range m {
		// fmt.Printf("Replace %s with %v \n", k, v)
		filebytes = bytes.Replace(filebytes, []byte(k), []byte(v), -1)
	}

	return string(filebytes)
}
