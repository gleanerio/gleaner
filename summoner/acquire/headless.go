package acquire

import (
	"bytes"
	"context"
	"crypto/sha1"
	"fmt"
	"log"
	"sync"

	"earthcube.org/Project418/gleaner/summoner/sitemaps"
	"earthcube.org/Project418/gleaner/summoner/utils"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/client"
	minio "github.com/minio/minio-go"
)

// Headless gets schema.org entries in sites that put the JSON-LD in dynamically with JS.
// It uses a chrome headless instance (which MUST BE RUNNING).
// TODO..  trap out error where headless is NOT running
func Headless(m map[string]sitemaps.URLSet, cs utils.Config) {
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

	// Create context and headless chrome instances
	ctxt, cancel := context.WithCancel(context.Background())
	defer cancel()
	c, err := chromedp.New(ctxt, chromedp.WithTargets(client.New().WatchPageTargets(ctxt)), chromedp.WithLog(log.Printf))
	if err != nil {
		log.Fatal(err)
	}

	// Set up some concurrency support
	semaphoreChan := make(chan struct{}, 1) // a blocking channel to keep concurrency under control
	defer close(semaphoreChan)
	wg := sync.WaitGroup{} // a wait group enables the main process a wait for goroutines to finish

	fmt.Println("headless before loops")
	fmt.Println(m)

	for k := range m {
		fmt.Printf("Act on URL's for %s", k)
		for i := range m[k].URL {

			wg.Add(1)

			urlloc := m[k].URL[i].Loc

			fmt.Println(urlloc)

			go func(i int, k string) {
				// block until the semaphore channel has room
				// this could also be moved out of the goroutine
				// which would make sense if the list is huge
				semaphoreChan <- struct{}{}

				// ru := urls[i].Loc
				var jsonld string
				err = c.Run(ctxt, domprocess(urlloc, &jsonld))
				if err != nil {
					log.Fatal(err)
				}

				// if jsonld != "" {
				// 	u, o, loaderror := LoadToMinio(jsonld, k, ru, minioClient, i)
				// 	if loaderror != nil {
				// 		log.Printf("Error loading to bucket: %s", ru)
				// 	}
				// 	fmt.Printf("Status: %v \n URL: %s \n ObjectName: %s \n schema.org --> \n", loaderror, u, o)
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
					log.Printf("#%d Uploaded Bucket:%s File:%s Size %d \n", i, bucketName, objectName, n)
					fmt.Printf("#%d Uploaded Bucket:%s File:%s Size %d \n", i, bucketName, objectName, n)

				}

				wg.Done() // tell the wait group that we be done

				fmt.Printf("#%d  got %s ", i, urlloc) // print an message containing the index (won't keep order)
				<-semaphoreChan                       // clear a spot in the semaphore channel
			}(i, k)

		}
	}

	wg.Wait()

}

func domprocess(targeturl string, res *string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(targeturl),
		// chromedp.Sleep(1 * time.Second),
		chromedp.Text(`#schemaorg`, res, chromedp.ByID),
	}
}
