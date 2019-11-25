package acquire

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"earthcube.org/Project418/gleaner/pkg/summoner/sitemaps"
	"github.com/chromedp/chromedp"
	minio "github.com/minio/minio-go"
)

// Headless gets schema.org entries in sites that put the JSON-LD in dynamically with JS.
// It uses a chrome headless instance (which MUST BE RUNNING).
// TODO..  trap out error where headless is NOT running
func Headless(minioClient *minio.Client, m map[string]sitemaps.URLSet) {
	// err := buildBuckets(minioClient, m) // TODO needs error obviously
	// if err != nil {
	// 	log.Printf("Gleaner bucket report:  %s", err)
	// }

	dt := getDebugURL()
	actxt, cancelActxt := chromedp.NewRemoteAllocator(context.Background(), dt)
	defer cancelActxt()

	ctxt, cancelCtxt := chromedp.NewContext(actxt) // create new tab
	defer cancelCtxt()                             // close tab afterwards

	// Create context and headless chrome instances
	// ctxt, cancel := chromedp.NewContext(context.Background())
	// defer cancel()

	// Set up some concurrency support
	semaphoreChan := make(chan struct{}, 1) // this HEADLESS is NOT thread safe yet!  Need  a blocking channel to keep concurrency under control
	defer close(semaphoreChan)
	wg := sync.WaitGroup{} // a wait group enables the main process a wait for goroutines to finish

	// log.Println("headless before loops")
	// log.Println(m)

	for k := range m {
		log.Printf("Headless chrome call to: %s", k)
		for i := range m[k].URL {

			wg.Add(1)
			urlloc := m[k].URL[i].Loc
			// log.Println(urlloc)

			go func(i int, k string) {
				semaphoreChan <- struct{}{}

				var jsonld string
				err := chromedp.Run(ctxt, domprocess(urlloc, &jsonld))
				if err != nil {
					log.Println(err)
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

					// Upload the zip file with FPutObject
					n, err := minioClient.PutObject(bucketName, objectName, b, int64(b.Len()), minio.PutObjectOptions{ContentType: contentType, UserMetadata: usermeta})
					if err != nil {
						log.Printf("%s", objectName)
						log.Println(err)
					}
					log.Printf("#%d Uploaded Bucket:%s File:%s Size %d \n", i, bucketName, objectName, n)
				}

				wg.Done() // tell the wait group that we be done

				log.Printf("#%d got %s ", i, urlloc) // print an message containing the index (won't keep order)
				<-semaphoreChan                      // clear a spot in the semaphore channel
			}(i, k)

		}
	}

	wg.Wait()

}

func getDebugURL() string {
	resp, err := http.Get("http://localhost:9222/json/version")
	if err != nil {
		log.Fatal(err)
	}

	var result map[string]interface{}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatal(err)
	}
	return result["webSocketDebuggerUrl"].(string)
}

func domprocess(targeturl string, res *string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(targeturl),
		chromedp.Sleep(5 * time.Second),
		chromedp.OuterHTML("html", res),
		// chromedp.Text(`#schemaorg`, res, chromedp.ByID),
	}
}
