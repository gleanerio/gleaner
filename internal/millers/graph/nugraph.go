package graph

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"strings"
	"sync"

	"earthcube.org/Project418/gleaner/internal/common"
	"earthcube.org/Project418/gleaner/internal/millers/millerutils"
	"earthcube.org/Project418/gleaner/pkg/utils"

	minio "github.com/minio/minio-go"
)

// NewGraphMillObjects doesn't marshal the objects to memory.  That doesn't work with
// millions of objects.  We have to stream them through...
func NewGraphMillObjects(mc *minio.Client, prefix string, cs utils.Config) {
	doneCh := make(chan struct{}) // Create a done channel to control 'ListObjectsV2' go routine.
	defer close(doneCh)           // Indicate to our routine to exit cleanly upon return.
	isRecursive := true
	bucketname := "gleaner-summoned"
	objectCh := mc.ListObjectsV2(bucketname, prefix, isRecursive, doneCh)

	semaphoreChan := make(chan struct{}, 20) // a blocking channel to keep concurrency under control
	defer close(semaphoreChan)
	wg := sync.WaitGroup{} // a wait group enables the main process a wait for goroutines to finish
	var gb common.Buffer
	k := 0

	for object := range objectCh {

		wg.Add(1)
		log.Printf("About to run #%d in a goroutine\n", k)

		go func(k int) {
			semaphoreChan <- struct{}{}

			if object.Err != nil {
				fmt.Println(object.Err)
			}

			fo, err := mc.GetObject(bucketname, object.Key, minio.GetObjectOptions{})
			if err != nil {
				fmt.Println(err)
			}
			oi, err := fo.Stat()
			if err != nil {
				log.Println("Issue with reading an object..  should I just fatal on this to make sure?")
			}
			urlval := ""
			sha1val := ""
			if len(oi.Metadata["X-Amz-Meta-Url"]) > 0 {
				urlval = oi.Metadata["X-Amz-Meta-Url"][0] // also have  X-Amz-Meta-Sha1
			}
			if len(oi.Metadata["X-Amz-Meta-Sha1"]) > 0 {
				sha1val = oi.Metadata["X-Amz-Meta-Sha1"][0]
			}
			buf := new(bytes.Buffer)
			buf.ReadFrom(fo)
			jld := buf.String() // Does a complete copy of the bytes in the buffer.

			status := millerutils.Jsl2graph(bucketname, object.Key, urlval, sha1val, jld, &gb)

			// fmt.Println(status)

			wg.Done() // tell the wait group that we be done!
			log.Printf("#%d wrote %d bytes", k, status)
			<-semaphoreChan
		}(k)

		k = k + 1

		// fmt.Printf("Processed object %d/%d with status: %d\n", n, c, status)

	}
	wg.Wait()

	log.Println(gb.Len())

	// STEP 1 clean triples  (split to two buffers..   good and bad)
	// STEP 1 covert good buffer NT to NQ (so I need the context from the config file to define the graph)
	var err error
	scanner := bufio.NewScanner(&gb) // rdf is already a pointer
	good := bytes.NewBuffer(make([]byte, 0))
	bad := bytes.NewBuffer(make([]byte, 0))
	for scanner.Scan() {
		if len(scanner.Text()) > 2 {
			nq, e := goodTriples(scanner.Text(), fmt.Sprintf("http://earthcube.org/%s", bucketname))
			if e == nil {
				_, err = good.Write([]byte(nq))
			}
			if e != nil {
				_, err = bad.Write([]byte(fmt.Sprintf("%s :Error: %s\n", strings.TrimSuffix(scanner.Text(), "\n"), e)))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// TODO  write both of these to the Minio system
	log.Println(good.Len())
	log.Println(bad.Len())

	// TODO: Can we clear up gb at this point if we use these good/bad buffers from here out?

	// write two object to S3; the quads and the error list
	flgood, err := millerutils.LoadToMinio(good.String(), "gleaner-milled", fmt.Sprintf("%s/%s.nq", cs.Gleaner.RunID, prefix), mc)
	if err != nil {
		log.Println("RDF file could not be written")
	} else {
		log.Printf("RDF file written len:%d\n", flgood)
	}
	if bad.Len() > 0 { // when the light is green, the trap is clean
		flbad, err := millerutils.LoadToMinio(bad.String(), "gleaner-milled", fmt.Sprintf("%s/%s_rdfErrors.txt", cs.Gleaner.RunID, prefix), mc)
		if err != nil {
			log.Println("RDF Error file could not be written")
		} else {
			log.Printf("RDF Error file written len:%d\n", flbad)
		}
	}

}
