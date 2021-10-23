package acquire

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/boltdb/bolt"
	"github.com/minio/minio-go/v7"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/viper"
)

// ResRetrieve is a function to pull down the data graphs at resources
func ResRetrieve(v1 *viper.Viper, mc *minio.Client, m map[string][]string, db *bolt.DB) {
	wg := sync.WaitGroup{}

	// Why do I pass the wg pointer?   Just make a new one
	// for each domain in getDomain and us this one here with a semaphore
	// to control the loop?
	for k := range m {
		// log.Printf("Queuing URLs for %s \n", k)
		go getDomain(v1, mc, m, k, &wg, db)
	}

	time.Sleep(2 * time.Second) // ?? why is this here?
	wg.Wait()
}

func getDomain(v1 *viper.Viper, mc *minio.Client, m map[string][]string, k string, wg *sync.WaitGroup, db *bolt.DB) {

	// read config file
	miniocfg := v1.GetStringMapString("minio")
	bucketName := miniocfg["bucket"] //   get the top level bucket for all of gleaner operations from config file

	// make the bucket (if it doesn't exist)
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(k))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	tc, err := Threadcount(v1)
	if err != nil {
		log.Println(err)
	}
	dt, err := Delayrequest(v1)
	if err != nil {
		log.Println(err)
	}

	if dt > 0 {
		tc = 1 // If the domain requests a delay between request, drop to single threaded and honor delay
	}

	log.Printf("Thread count %d delay %d\n", tc, dt)

	semaphoreChan := make(chan struct{}, tc) // a blocking channel to keep concurrency under control
	defer close(semaphoreChan)
	lwg := sync.WaitGroup{}

	wg.Add(1)       // wg from the calling function
	defer wg.Done() // tell the wait group that we be done

	count := len(m[k])
	bar := progressbar.Default(int64(count))

	var (
		buf    bytes.Buffer
		logger = log.New(&buf, "logger: ", log.Lshortfile)
	)

	// var client http.Client

	// we actually go get the URLs now
	for i := range m[k] {
		lwg.Add(1)
		urlloc := m[k][i]

		// TODO / WARNING for large site we can exhaust memory with just the creation of the
		// go routines. 1 million =~ 4 GB  So we need to control how many routines we
		// make too..  reference https://github.com/mr51m0n/gorc (but look for someting in the core
		// library too)

		// log.Println(urlloc)

		go func(i int, k string) {
			semaphoreChan <- struct{}{}

			// log.Println(urlloc)

			urlloc = strings.ReplaceAll(urlloc, " ", "")
			urlloc = strings.ReplaceAll(urlloc, "\n", "")

			var client http.Client // why do I make this here..  can I use 1 client?  move up in the loop
			req, err := http.NewRequest("GET", urlloc, nil)
			if err != nil {
				log.Println(err)
				logger.Printf("#%d error on %s : %s  ", i, urlloc, err) // print an message containing the index (won't keep order)
			}
			req.Header.Set("User-Agent", "EarthCube_DataBot/1.0")
			req.Header.Set("Accept", "application/ld+json, text/html")

			resp, err := client.Do(req)
			if err != nil {
				logger.Printf("#%d error on %s : %s  ", i, urlloc, err) // print an message containing the index (won't keep order)
				lwg.Done()                                              // tell the wait group that we be done
				<-semaphoreChan
				return
			}
			defer resp.Body.Close()

			doc, err := goquery.NewDocumentFromResponse(resp)
			if err != nil {
				logger.Printf("#%d error on %s : %s  ", i, urlloc, err) // print an message containing the index (won't keep order)
				lwg.Done()                                              // tell the wait group that we be done
				<-semaphoreChan
				return
			}

			var jsonlds []string
			var contentTypeHeader = resp.Header["Content-Type"]

			// The URL is sending back JSON-LD correctly as application/ld+json
			// this should not be here IMHO, but need to support people not setting proper header value
			// The URL is sending back JSON-LD but incorrectly sending as application/json
			if contains(contentTypeHeader, "application/ld+json") || contains(contentTypeHeader, "application/json") {
				jsonlds, err = addToJsonListIfValid(v1, jsonlds, doc.Text())
				if err != nil {
					logger.Printf("Error processing json response from %s: %s", urlloc, err)
				}
				// look in the HTML page for <script type=application/ld+json
			} else {
				doc.Find("script").Each(func(i int, s *goquery.Selection) {
					val, _ := s.Attr("type")
					if val == "application/ld+json" {
						jsonlds, err = addToJsonListIfValid(v1, jsonlds, s.Text())
						if err != nil {
							logger.Printf("Error processing script tag in %s: %s", urlloc, err)
						}
					}
				})
			}

			// For incremental indexing I want to know every URL I visit regardless
			// if there is a valid JSON-LD document or not.   For "full" indexing we
			// visit ALL URLs.  However, many will not have JSON-LD, so let's also record
			// and avoid those during incremental calls.

			// even is no JSON-LD packages found, record the event
			if len(jsonlds) < 1 {
				db.Update(func(tx *bolt.Tx) error {
					b := tx.Bucket([]byte(k))
					err := b.Put([]byte(urlloc), []byte(fmt.Sprintf("NILL: %s", urlloc))) // no JOSN-LD found at this URL
					return err
				})
			}

			for _, jsonld := range jsonlds {
				if jsonld != "" { // traps out the root domain...   should do this different
					logger.Printf("#%d Uploading", i)
					sha, err := Upload(v1, mc, logger, bucketName, k, urlloc, jsonld)
					if err != nil {
						logger.Printf("Error uploading jsonld to object store: %s: %s", urlloc, err)
					}
					// TODO  Is here where to add an entry to the KV store
					db.Update(func(tx *bolt.Tx) error {
						b := tx.Bucket([]byte(k))
						err := b.Put([]byte(urlloc), []byte(sha))
						return err
					})
				} else {
					logger.Printf("Empty JSON-LD document found. Continuing.")
					// TODO  Is here where to add an entry to the KV store
					db.Update(func(tx *bolt.Tx) error {
						b := tx.Bucket([]byte(k))
						err := b.Put([]byte(urlloc), []byte(fmt.Sprintf("NULL: %s", urlloc))) // no JOSN-LD found at this URL
						return err
					})
				}
			}

			bar.Add(1)                                       // bar.Incr()
			logger.Printf("#%d thread for %s ", i, urlloc)   // print an message containing the index (won't keep order)
			time.Sleep(time.Duration(dt) * time.Millisecond) // sleep a bit if directed to by the provider

			lwg.Done()

			<-semaphoreChan // clear a spot in the semaphore channel for the next indexing event
		}(i, k)

	}

	lwg.Wait()

	// TODO write this to minio in the run ID bucket
	// return the logger buffer or write to a mutex locked bytes buffer
	f, err := os.Create(fmt.Sprintf("./%s.log", k))
	if err != nil {
		log.Println("Error writing a file")
	}

	w := bufio.NewWriter(f)
	_, err = w.WriteString(buf.String())
	if err != nil {
		log.Println("Error writing a file")
	}
	w.Flush()

}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		// if a == str {
		if strings.Contains(a, str) {
			return true
		}
	}
	return false
}

func rightPad2Len(s string, padStr string, overallLen int) string {
	var padCountInt int
	padCountInt = 1 + ((overallLen - len(padStr)) / len(padStr))
	var retStr = s + strings.Repeat(padStr, padCountInt)
	return retStr[:overallLen]
}
