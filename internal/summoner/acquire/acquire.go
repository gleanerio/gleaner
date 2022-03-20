package acquire

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	configTypes "github.com/gleanerio/gleaner/internal/config"

	"github.com/PuerkitoBio/goquery"
	"github.com/minio/minio-go/v7"
	"github.com/samclarke/robotstxt"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/viper"
	bolt "go.etcd.io/bbolt"
)

const EarthCubeAgent = "EarthCube_DataBot/1.0"

// ResRetrieve is a function to pull down the data graphs at resources
func ResRetrieve(v1 *viper.Viper, mc *minio.Client, m map[string][]string, db *bolt.DB) {
	wg := sync.WaitGroup{}

	// Why do I pass the wg pointer?   Just make a new one
	// for each domain in getDomain and us this one here with a semaphore
	// to control the loop?
	for domain, urls := range m {
		log.Printf("Queuing URLs for %s \n", domain)
		go getDomain(v1, mc, urls, domain, &wg, db)
	}

	time.Sleep(2 * time.Second) // ?? why is this here?
	wg.Wait()
}

func getConfig(v1 *viper.Viper) (string, int, int64, error) {
	bucketName, err := configTypes.GetBucketName(v1)
	if err != nil {
		return bucketName, 0, 0, err
	}

	var mcfg configTypes.Summoner
	mcfg, err = configTypes.ReadSummmonerConfig(v1.Sub("summoner"))
	tc := mcfg.Threads
	delay := mcfg.Delay

	if err != nil {
		return bucketName, tc, delay, err
	}

	if delay != 0 {
		tc = 1
	}

	log.Printf("Thread count %d delay %d\n", tc, delay)
	return bucketName, tc, delay, nil
}

func getRobotsForDomain(v1 *viper.Viper, sourceName string) (*robotstxt.RobotsTxt, error) {
	// first get the domain url for our source
	sourcesConfig, err := configTypes.GetSources(v1)
	domain, err := configTypes.GetSourceByName(sourcesConfig, sourceName)
	if err != nil {
		log.Printf("error getting domain url for %s : %s  ", sourceName, err)
		return nil, err
	}

	var robotsUrl string
	if domain.SourceType == "robots" {
		robotsUrl = domain.URL
	} else {
		robotsUrl = domain.Domain + "/robots.txt"
	}

	robots, err := getRobotsTxt(robotsUrl)
	if err != nil {
		log.Printf("error getting robots.txt for %s : %s  ", sourceName, err)
		return nil, err
	}

	return robots, nil
}

func getDomain(v1 *viper.Viper, mc *minio.Client, urls []string, sourceName string, wg *sync.WaitGroup, db *bolt.DB) {

	// make the bucket (if it doesn't exist)
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(sourceName))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	bucketName, tc, delay, err := getConfig(v1)
	if err != nil {
		log.Panic("Error reading config file", err)
	}

	var client http.Client

	robots, err := getRobotsForDomain(v1, sourceName)

	if err != nil {
		log.Printf("Error getting robots.txt for %s; continuing without it.", sourceName)
	}

	// Look at the crawl delay from this domain's robots.txt, if we can, and one exists.
	if robots != nil {
		// this is a time.Duration, which is in nanoseconds, because of COURSE it is, but we want milliseconds
		crawlDelay := int64(robots.CrawlDelay(EarthCubeAgent) / time.Millisecond)
		log.Printf("Crawl Delay specified by robots.txt for %s: %d", sourceName, crawlDelay)

		// If our default delay is less than what is set there, bump up the delay for this
		// domain to respect the robots.txt setting.
		if delay < crawlDelay {
			delay = crawlDelay
			tc = 1 // any delay means going down to one thread.
		}
	}

	semaphoreChan := make(chan struct{}, tc) // a blocking channel to keep concurrency under control
	defer close(semaphoreChan)
	lwg := sync.WaitGroup{}

	wg.Add(1)       // wg from the calling function
	defer wg.Done() // tell the wait group that we be done

	count := len(urls)
	bar := progressbar.Default(int64(count))

	var (
		buf    bytes.Buffer
		logger = log.New(&buf, "logger: ", log.Lshortfile)
	)

	// we actually go get the URLs now
	for i := range urls {
		lwg.Add(1)
		urlloc := urls[i]

		// TODO / WARNING for large site we can exhaust memory with just the creation of the
		// go routines. 1 million =~ 4 GB  So we need to control how many routines we
		// make too..  reference https://github.com/mr51m0n/gorc (but look for someting in the core
		// library too)

		go func(i int, sourceName string) {
			semaphoreChan <- struct{}{}
			logger.Println("Indexing", urlloc)

			urlloc = strings.ReplaceAll(urlloc, " ", "")
			urlloc = strings.ReplaceAll(urlloc, "\n", "")

			if robots != nil {
				allowed, err := robots.IsAllowed(EarthCubeAgent, urlloc)
				if !allowed {
					logger.Printf("Declining to index %s because it is disallowed by robots.txt. Error information, if any: %s", urlloc, err)
					lwg.Done() // tell the wait group that we be done
					<-semaphoreChan
					return
				}
			}

			req, err := http.NewRequest("GET", urlloc, nil)
			if err != nil {
				log.Println(err)
				logger.Printf("#%d error on %s : %s  ", i, urlloc, err) // print an message containing the index (won't keep order)
			}
			req.Header.Set("User-Agent", EarthCubeAgent)
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

			// if
			// The URL is sending back JSON-LD correctly as application/ld+json
			// this should not be here IMHO, but need to support people not setting proper header value
			// The URL is sending back JSON-LD but incorrectly sending as application/json
			if contains(contentTypeHeader, "application/ld+json") || contains(contentTypeHeader, "application/json") || fileExtensionIsJson(urlloc) {
				logger.Printf("%s as %s", urlloc, contentTypeHeader)
				jsonlds, err = addToJsonListIfValid(v1, jsonlds, doc.Text())
				if err != nil {
					logger.Printf("Error processing json response from %s: %s", urlloc, err)
				}
				// look in the HTML page for <script type=application/ld+json>
			} else {
				doc.Find("script[type='application/ld+json']").Each(func(i int, s *goquery.Selection) {
					jsonlds, err = addToJsonListIfValid(v1, jsonlds, s.Text())
					if err != nil {
						logger.Printf("Error processing script tag in %s: %s", urlloc, err)
					}
				})
			}

			// For incremental indexing I want to know every URL I visit regardless
			// if there is a valid JSON-LD document or not.   For "full" indexing we
			// visit ALL URLs.  However, many will not have JSON-LD, so let's also record
			// and avoid those during incremental calls.

			// even is no JSON-LD packages found, record the event of checking this URL
			if len(jsonlds) < 1 {
				// TODO is her where I then try headless, and scope the following for into an else?
				log.Printf("Direct access failed, trying headless for  %s ", urlloc)
				err := PageRender(v1, mc, logger, 60*time.Second, urlloc, sourceName, db) // TODO make delay configurable
				if err != nil {
					logger.Printf("PageRender %s :: %s", urlloc, err)
				}
				db.Update(func(tx *bolt.Tx) error {
					b := tx.Bucket([]byte(sourceName))
					err := b.Put([]byte(urlloc), []byte(fmt.Sprintf("NILL: %s", urlloc))) // no JOSN-LD found at this URL
					return err
				})
				if err != nil {
					logger.Printf("DB Update %s :: %s", urlloc, err)
				}

			} //else {
			//log.Printf("Direct access worked for  %s ", urlloc)
			//}

			for i, jsonld := range jsonlds {
				if jsonld != "" { // traps out the root domain...   should do this different
					logger.Printf("#%d Uploading", i)
					sha, err := Upload(v1, mc, logger, bucketName, sourceName, urlloc, jsonld)
					if err != nil {
						logger.Printf("Error uploading jsonld to object store: %s: %s", urlloc, err)
					}
					// TODO  Is here where to add an entry to the KV store
					db.Update(func(tx *bolt.Tx) error {
						b := tx.Bucket([]byte(sourceName))
						err := b.Put([]byte(urlloc), []byte(sha))
						return err
					})
				} else {
					logger.Printf("Empty JSON-LD document found. Continuing.")
					// TODO  Is here where to add an entry to the KV store
					db.Update(func(tx *bolt.Tx) error {
						b := tx.Bucket([]byte(sourceName))
						err := b.Put([]byte(urlloc), []byte(fmt.Sprintf("NULL: %s", urlloc))) // no JOSN-LD found at this URL
						return err
					})
				}
			}

			bar.Add(1)                                          // bar.Incr()
			logger.Printf("#%d thread for %s ", i, urlloc)      // print an message containing the index (won't keep order)
			time.Sleep(time.Duration(delay) * time.Millisecond) // sleep a bit if directed to by the provider

			lwg.Done()

			<-semaphoreChan // clear a spot in the semaphore channel for the next indexing event
		}(i, sourceName)

	}

	lwg.Wait()

	// TODO write this to minio in the run ID bucket
	// return the logger buffer or write to a mutex locked bytes buffer
	f, err := os.Create(fmt.Sprintf("./%s.log", sourceName))
	if err != nil {
		log.Println("Error writing a file")
	}

	w := bufio.NewWriter(f)
	bc, err := w.WriteString(buf.String())
	if err != nil {
		log.Println("Error writing a file")
	}
	w.Flush()
	log.Printf("Wrote log size %d", bc)

}

func contains(arr []string, str string) bool {
	for _, a := range arr {

		if strings.Contains(a, str) {
			return true
		}
	}
	return false
}

func fileExtensionIsJson(rawUrl string) bool {
	u, _ := url.Parse(rawUrl)
	if strings.HasSuffix(u.Path, ".json") || strings.HasSuffix(u.Path, ".jsonld") {
		return true
	}
	return false
}
