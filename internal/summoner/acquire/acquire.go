package acquire

import (
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	configTypes "github.com/gleanerio/gleaner/internal/config"

	"github.com/PuerkitoBio/goquery"
	"github.com/minio/minio-go/v7"
	"github.com/samclarke/robotstxt"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/viper"
)

const EarthCubeAgent = "EarthCube_DataBot/1.0"

// ResRetrieve is a function to pull down the data graphs at resources
func ResRetrieve(v1 *viper.Viper, mc *minio.Client, m map[string][]string) {
	wg := sync.WaitGroup{}

	// Why do I pass the wg pointer?   Just make a new one
	// for each domain in getDomain and us this one here with a semaphore
	// to control the loop?
	for domain, urls := range m {
		log.Printf("Queuing URLs for %s \n", domain)
		go getDomain(v1, mc, urls, domain, &wg)
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
	log.Println(robots)
	if err != nil {
		log.Printf("error getting robots.txt for %s : %s  ", sourceName, err)
		return nil, err
	}

	return robots, nil
}

func getDomain(v1 *viper.Viper, mc *minio.Client, urls []string, sourceName string, wg *sync.WaitGroup) {
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

	// we actually go get the URLs now
	for i := range urls {
		lwg.Add(1)
		urlloc := urls[i]

		// TODO / WARNING for large site we can exhaust memory with just the creation of the
		// go routines. 1 million =~ 4 GB  So we need to control how many routines we
		// make too..  reference https://github.com/mr51m0n/gorc (but look for something in the core
		// library too)

		go func(i int, sourceName string) {
			semaphoreChan <- struct{}{}
			log.Println("Indexing", urlloc)

			urlloc = strings.ReplaceAll(urlloc, " ", "")
			urlloc = strings.ReplaceAll(urlloc, "\n", "")

			// TODO fix robot opt in here
			//if robots != nil {
			//allowed, err := robots.IsAllowed(EarthCubeAgent, urlloc)
			//if !allowed {
			//log.Printf("Declining to index %s because it is disallowed by robots.txt. Error information, if any: %s", urlloc, err)
			//lwg.Done() // tell the wait group that we be done
			//<-semaphoreChan
			//return
			//}
			//}

			req, err := http.NewRequest("GET", urlloc, nil)
			if err != nil {
				log.Println(err)
				log.Printf("#%d error on %s : %s  ", i, urlloc, err) // print an message containing the index (won't keep order)
			}
			req.Header.Set("User-Agent", EarthCubeAgent)
			req.Header.Set("Accept", "application/ld+json, text/html")

			resp, err := client.Do(req)
			if err != nil {
				log.Printf("#%d error on %s : %s  ", i, urlloc, err) // print an message containing the index (won't keep order)
				lwg.Done()                                           // tell the wait group that we be done
				<-semaphoreChan
				return
			}
			defer resp.Body.Close()

			doc, err := goquery.NewDocumentFromResponse(resp)
			if err != nil {
				log.Printf("#%d error on %s : %s  ", i, urlloc, err) // print an message containing the index (won't keep order)
				lwg.Done()                                           // tell the wait group that we be done
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
				log.Printf("%s as %s", urlloc, contentTypeHeader)
				jsonlds, err = addToJsonListIfValid(v1, jsonlds, doc.Text())
				if err != nil {
					log.Printf("Error processing json response from %s: %s", urlloc, err)
				}
				// look in the HTML page for <script type=application/ld+json>
			} else {
				doc.Find("script[type='application/ld+json']").Each(func(i int, s *goquery.Selection) {
					jsonlds, err = addToJsonListIfValid(v1, jsonlds, s.Text())
					if err != nil {
						log.Printf("Error processing script tag in %s: %s", urlloc, err)
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
				err := PageRender(v1, mc, 60*time.Second, urlloc, sourceName) // TODO make delay configurable
				if err != nil {
					log.Printf("PageRender %s :: %s", urlloc, err)
				}

			} //else {
			//log.Printf("Direct access worked for  %s ", urlloc)
			//}

			for i, jsonld := range jsonlds {
				if jsonld != "" { // traps out the root domain...   should do this different
					log.Printf("#%d Uploading", i)
					_, err := Upload(v1, mc, bucketName, sourceName, urlloc, jsonld)
					if err != nil {
						log.Printf("Error uploading jsonld to object store: %s: %s", urlloc, err)
					}

				} else {
					log.Printf("Empty JSON-LD document found. Continuing.")

				}
			}

			bar.Add(1)                                          // bar.Incr()
			log.Printf("#%d thread for %s ", i, urlloc)         // print an message containing the index (won't keep order)
			time.Sleep(time.Duration(delay) * time.Millisecond) // sleep a bit if directed to by the provider

			lwg.Done()

			<-semaphoreChan // clear a spot in the semaphore channel for the next indexing event
		}(i, sourceName)

	}

	lwg.Wait()

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
