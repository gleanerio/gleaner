package acquire

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"
	"net/http"
	"sync"

	"earthcube.org/Project418/gleaner/summoner/sitemaps"
	"earthcube.org/Project418/gleaner/summoner/utils"
	"github.com/PuerkitoBio/goquery"
	"github.com/minio/minio-go"
)

// ResRetrieve
func ResRetrieve(m map[string]sitemaps.URLSet, cs utils.Config) {

	// Set up minio and initialize client
	endpoint := cs.Minio.Endpoint
	accessKeyID := cs.Minio.AccessKeyID
	secretAccessKey := cs.Minio.SecretAccessKey
	useSSL := false
	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		log.Fatalln(err)
	}
	buildBuckets(minioClient, m) // TODO needs error obviously

	// set up some concurrency support
	semaphoreChan := make(chan struct{}, 15) // a blocking channel to keep concurrency under control
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
				}
				defer resp.Body.Close()

				doc, err := goquery.NewDocumentFromResponse(resp)
				if err != nil {
					log.Printf("error %v", err)
				}

				// TODO Version that just looks for script type application/ld+json
				// this will look for ALL nodes in the doc that match, there may be more than one
				var jsonld string
				if err == nil {
					doc.Find("script").Each(func(i int, s *goquery.Selection) {
						val, _ := s.Attr("type")
						if val == "application/ld+json" {
							err = isValid(s.Text())
							if err != nil {
								log.Printf("ERROR: At %s JSON-LD is NOT valid: %s", urlloc, err)
							}
							jsonld = s.Text()
						}
					})
				}

				// if jsonld != "" {
				// 	u, o, err := LoadToMinio(jsonld, k, urlloc, minioClient, i)
				// 	if err != nil {
				// 		log.Printf("Error loading to bucket: %s", urlloc)
				// 	}
				// 	fmt.Printf("Status: %v \n URL: %s \n ObjectName: %s \n schema.org --> \n", err, u, o)

				// }

				if jsonld != "" { // traps out the root domain...   should do this different
					// get sha1 of the JSONLD..  it's a nice ID
					h := sha1.New()
					h.Write([]byte(jsonld))
					bs := h.Sum(nil)
					bss := fmt.Sprintf("%x", bs) // better way to convert bs hex string to string?

					// objectName := fmt.Sprintf("%s/%s.jsonld", up.Path, bss)
					objectName := fmt.Sprintf("%s.jsonld", bss)
					contentType := "application/ld+json"
					b := bytes.NewBufferString(jsonld)

					usermeta := make(map[string]string) // what do I want to know?
					usermeta["url"] = urlloc
					usermeta["sha1"] = bss
					bucketName := k

					// Upload the zip file with FPutObject
					n, err := minioClient.PutObject(bucketName, objectName, b, int64(b.Len()), minio.PutObjectOptions{ContentType: contentType, UserMetadata: usermeta})
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
