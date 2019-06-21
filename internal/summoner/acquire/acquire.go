package acquire

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"earthcube.org/Project418/gleaner/pkg/summoner/sitemaps"
	"earthcube.org/Project418/gleaner/pkg/utils"
	"github.com/PuerkitoBio/goquery"
	"github.com/kazarena/json-gold/ld"
	"github.com/minio/minio-go"
)

// ResRetrieve is a function to pull down the data graphs at resources
func ResRetrieve(mc *minio.Client, m map[string]sitemaps.URLSet, cs utils.Config) {
	err := buildBuckets(mc, m) // TODO needs error obviously
	if err != nil {
		log.Printf("Gleaner bucket report:  %s", err)
	}

	// set up some concurrency support
	semaphoreChan := make(chan struct{}, 5) // a blocking channel to keep concurrency under control
	defer close(semaphoreChan)
	wg := sync.WaitGroup{} // a wait group enables the main process a wait for goroutines to finish

	for k := range m {
		log.Printf("Act on URL's for %s", k)
		for i := range m[k].URL {

			wg.Add(1)

			// log.Printf("----> %s", m[k].URL[i].Loc)
			urlloc := m[k].URL[i].Loc

			go func(i int, k string) {
				// block until the semaphore channel has room
				// this could also be moved out of the goroutine
				// which would make sense if the list is huge
				semaphoreChan <- struct{}{}

				var client http.Client
				req, err := http.NewRequest("GET", urlloc, nil)
				if err != nil {
					// not even being able to make a req instance..  might be a fatal thing?
					log.Printf("------ error making request------ \n %s", err)
				}

				req.Header.Set("User-Agent", "EarthCube_DataBot/1.0")

				resp, err := client.Do(req)
				if err != nil {
					log.Printf("Error reading location: %s", err)
					log.Printf("#%d error on %s ", i, urlloc) // print an message containing the index (won't keep order)
					wg.Done()                                 // tell the wait group that we be done
					<-semaphoreChan
					return
				}
				defer resp.Body.Close()

				doc, err := goquery.NewDocumentFromResponse(resp)
				if err != nil {
					log.Printf("Error doc from resp: %v", err)
					log.Printf("#%d error on %s ", i, urlloc) // print an message containing the index (won't keep order)
					wg.Done()                                 // tell the wait group that we be done
					<-semaphoreChan
					return
				}

				// TODO Version that just looks for script type application/ld+json
				// this will look for ALL nodes in the doc that match, there may be more than one
				var jsonld string
				if err == nil {
					doc.Find("script").Each(func(i int, s *goquery.Selection) {
						val, _ := s.Attr("type")
						if val == "application/ld+json" {
							action, err := isValid(s.Text())
							if err != nil {
								log.Printf("ERROR: URL: %s Action: %s  Error: %s", urlloc, action, err)
							}
							jsonld = s.Text()
						}
					})
				}

				if jsonld != "" { // traps out the root domain...   should do this different
					// get sha1 of the JSONLD..  it's a nice ID
					h := sha1.New()
					h.Write([]byte(jsonld))
					bs := h.Sum(nil)
					bss := fmt.Sprintf("%x", bs) // better way to convert bs hex string to string?

					// objectName := fmt.Sprintf("%s/%s.jsonld", up.Path, bss)
					// objectName := fmt.Sprintf("%s.jsonld", bss)
					objectName := fmt.Sprintf("%s/%s.jsonld", k, bss)
					contentType := "application/ld+json"
					b := bytes.NewBufferString(jsonld)

					usermeta := make(map[string]string) // what do I want to know?
					usermeta["url"] = urlloc
					usermeta["sha1"] = bss
					bucketName := "gleaner-summoned"
					//bucketName := fmt.Sprintf("gleaner-summoned/%s", k) // old was just k

					// Upload the file with FPutObject
					n, err := mc.PutObject(bucketName, objectName, b, int64(b.Len()), minio.PutObjectOptions{ContentType: contentType, UserMetadata: usermeta})
					if err != nil {
						log.Printf("%s", objectName)
						log.Fatalln(err)
					}
					log.Printf("#%d Uploaded Bucket:%s File:%s Size %d\n", i, bucketName, objectName, n)
				}

				log.Printf("#%d acting on %s ", i, urlloc) // print an message containing the index (won't keep order)
				wg.Done()                                  // tell the wait group that we be done
				<-semaphoreChan                            // clear a spot in the semaphore channel
			}(i, k)

		}
	}

	wg.Wait() // wait for all the goroutines to be done
}

func isValid(jsonld string) (string, error) {
	proc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")
	options.Format = "application/nquads"

	var myInterface interface{}
	action := ""

	err := json.Unmarshal([]byte(jsonld), &myInterface)
	if err != nil {
		action = "json.Unmarshal call"
		return "", err
	}

	_, err = proc.ToRDF(myInterface, options) // returns triples but toss them, just validating
	if err != nil {
		action = "JSON-LD to RDF call"
		return "", err
	}

	return action, err
}
