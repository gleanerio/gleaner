package acquire

import (
	"fmt"
	"github.com/gleanerio/gleaner/internal/common"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	configTypes "github.com/gleanerio/gleaner/internal/config"
	"github.com/gleanerio/gleaner/internal/summoner/buckets"
	"github.com/minio/minio-go/v7"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	bolt "go.etcd.io/bbolt"
)

const EarthCubeAgent = "EarthCube_DataBot/1.0"

// ResRetrieve is a function to pull down the data graphs at resources
func ResRetrieve(v1 *viper.Viper, mc *minio.Client, m map[string][]string, db *bolt.DB, runStats *common.RunStats) {
	wg := sync.WaitGroup{}

	bucketName, err := configTypes.GetBucketName(v1)
	if err != nil {
		log.Error("Error getting bucket name in summoner ResRetrieve, needed to archive objects:", err)
	}

	// Why do I pass the wg pointer?   Just make a new one
	// for each domain in getDomain and us this one here with a semaphore
	// to control the loop?
	for domain, urls := range m {
		r := runStats.Add(domain)
		r.Set(common.Count, len(urls))
		r.Set(common.HttpError, 0)
		r.Set(common.Issues, 0)
		r.Set(common.Summoned, 0)
		fmt.Println("Archiving current objects for ", domain)
		log.Info("Archiving current objects for ", domain)

		// Prior to doing the indexing, archive the current object.  This could be a boolean option pulled from config to do this too.
		buckets.Copy(mc, bucketName, fmt.Sprintf("summoned/%s", domain), fmt.Sprintf("archive/%s", domain))
		buckets.Remove(mc, bucketName, fmt.Sprintf("summoned/%s", domain))

		log.Info("Queuing URLs for ", domain)

		repologger, err := common.LogIssues(v1, domain)
		if err != nil {
			log.Error("Error creating a logger for a repository", err)
		} else {
			repologger.Info("Queuing URLs for ", domain)
			repologger.Info("URL Count ", len(urls))
		}
		wg.Add(1)
		//go getDomain(v1, mc, urls, domain, &wg, db)
		go getDomain(v1, mc, urls, domain, &wg, db, repologger, r)
	}

	wg.Wait()
}

func getConfig(v1 *viper.Viper, sourceName string) (string, int, int64, error) {
	bucketName, err := configTypes.GetBucketName(v1)
	if err != nil {
		return bucketName, 0, 0, err
	}

	var mcfg configTypes.Summoner
	mcfg, err = configTypes.ReadSummmonerConfig(v1.Sub("summoner"))

	if err != nil {
		return bucketName, 0, 0, err
	}
	// Set default thread counts and global delay
	tc := mcfg.Threads
	delay := mcfg.Delay

	if delay != 0 {
		tc = 1
	}

	// look for a domain specific override crawl delay
	sources, err := configTypes.GetSources(v1)
	source, err := configTypes.GetSourceByName(sources, sourceName)

	if err != nil {
		return bucketName, tc, delay, err
	}

	if source.Delay != 0 && source.Delay > delay {
		delay = source.Delay
		tc = 1
		log.Info("Crawl delay set to ", delay, " for ", sourceName)
	}

	log.Info("Thread count ", tc, " delay ", delay)
	return bucketName, tc, delay, nil
}

func getDomain(v1 *viper.Viper, mc *minio.Client, urls []string, sourceName string,
	wg *sync.WaitGroup, db *bolt.DB, repologger *log.Logger, repoStats *common.RepoStats) {
	//func getDomain(v1 *viper.Viper, mc *minio.Client, urls []string, sourceName string, wg *sync.WaitGroup, db *bolt.DB) {

	// make the bucket (if it doesn't exist)
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(sourceName))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	// TODO..   I have this.   a merge removed it ..  I put it back..   what is the function of this one?
	//wg.Add(1) // wg from the calling function

	bucketName, tc, delay, err := getConfig(v1, sourceName)
	if err != nil {
		// trying to read a source, so let's not kill everything with a panic/fatal
		log.Error("Error reading config file ", err)
		repologger.Error("Error reading config file ", err)
	}

	var client http.Client

	semaphoreChan := make(chan struct{}, tc) // a blocking channel to keep concurrency under control
	lwg := sync.WaitGroup{}

	defer func() {
		lwg.Wait()
		wg.Done()
		close(semaphoreChan)
	}()

	count := len(urls)
	bar := progressbar.Default(int64(count))

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

			repologger.Trace("Indexing", urlloc)
			log.Debug("Indexing ", urlloc)

			req, err := http.NewRequest("GET", urlloc, nil)
			if err != nil {
				log.Error(i, err, urlloc)
			}
			req.Header.Set("User-Agent", EarthCubeAgent)
			req.Header.Set("Accept", "application/ld+json, text/html")

			resp, err := client.Do(req)
			if err != nil {
				log.Error("#", i, " error on ", urlloc, err) // print an message containing the index (won't keep order)
				repologger.WithFields(log.Fields{"url": urlloc}).Error(err)
				lwg.Done() // tell the wait group that we be done
				<-semaphoreChan
				return
			}
			defer resp.Body.Close()

			doc, err := goquery.NewDocumentFromResponse(resp)
			if err != nil {
				log.Error("#", i, " error on ", urlloc, err) // print an message containing the index (won't keep order)
				repoStats.Inc(common.Issues)
				lwg.Done() // tell the wait group that we be done
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
				repologger.WithFields(log.Fields{"url": urlloc, "contentType": "json or ld_json"}).Debug()
				log.WithFields(log.Fields{"url": urlloc, "contentType": "json or ld_json"}).Debug(urlloc, " as ", contentTypeHeader)

				jsonlds, err = addToJsonListIfValid(v1, jsonlds, doc.Text())
				if err != nil {
					log.WithFields(log.Fields{"url": urlloc, "contentType": "json or ld_json"}).Error("Error processing json response from ", urlloc, err)
					repologger.WithFields(log.Fields{"url": urlloc, "contentType": "json or ld_json"}).Error(err)
				}
				// look in the HTML page for <script type=application/ld+json>
			} else {
				doc.Find("script[type='application/ld+json']").Each(func(i int, s *goquery.Selection) {
					jsonlds, err = addToJsonListIfValid(v1, jsonlds, s.Text())
					repologger.WithFields(log.Fields{"url": urlloc, "contentType": "script[type='application/ld+json']"}).Info()
					if err != nil {
						log.WithFields(log.Fields{"url": urlloc, "contentType": "script[type='application/ld+json']"}).Error("Error processing script tag in ", urlloc, err)
						repologger.WithFields(log.Fields{"url": urlloc, "contentType": "script[type='application/ld+json']"}).Error(err)
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
				log.WithFields(log.Fields{"url": urlloc, "contentType": "Direct access failed, trying headless']"}).Info("Direct access failed, trying headless for ", urlloc)
				repologger.WithFields(log.Fields{"url": urlloc, "contentType": "Direct access failed, trying headless']"}).Error() // this needs to go into the issues file
				err := PageRender(v1, mc, 60*time.Second, urlloc, sourceName, db, repologger, repoStats)                           // TODO make delay configurable

				if err != nil {
					log.WithFields(log.Fields{"url": urlloc, "issue": "converting json ld"}).Error("PageRender ", urlloc, "::", err)
					repologger.WithFields(log.Fields{"url": urlloc, "issue": "converting json ld"}).Error(err)
				}
				db.Update(func(tx *bolt.Tx) error {
					b := tx.Bucket([]byte(sourceName))
					err := b.Put([]byte(urlloc), []byte(fmt.Sprintf("NILL: %s", urlloc))) // no JOSN-LD found at this URL
					return err
				})
				if err != nil {
					log.Error("DB Update", urlloc, "::", err)
				}

			} else {
				log.WithFields(log.Fields{"url": urlloc, "issue": "Direct access worked"}).Trace("Direct access worked for ", urlloc)
				repologger.WithFields(log.Fields{"url": urlloc, "issue": "Direct access worked"}).Trace()
				repoStats.Inc(common.Summoned)
			}

			for i, jsonld := range jsonlds {
				if jsonld != "" { // traps out the root domain...   should do this different
					log.WithFields(log.Fields{"url": urlloc, "issue": "Uploading"}).Trace("#", i, "Uploading ")
					repologger.WithFields(log.Fields{"url": urlloc, "issue": "Uploading"}).Trace()
					sha, err := Upload(v1, mc, bucketName, sourceName, urlloc, jsonld)
					if err != nil {
						log.WithFields(log.Fields{"url": urlloc, "sha": sha, "issue": "Error uploading jsonld to object store"}).Error("Error uploading jsonld to object store: ", urlloc, err)
						repologger.WithFields(log.Fields{"url": urlloc, "sha": sha, "issue": "Error uploading jsonld to object store"}).Error(err)
						repoStats.Inc(common.StoreError)
					} else {
						repologger.WithFields(log.Fields{"url": urlloc, "sha": sha, "issue": "Uploaded to object store"}).Trace(err)
						log.WithFields(log.Fields{"url": urlloc, "sha": sha, "issue": "Uploaded to object store"}).Info("Successfully put ", sha, " in summoned bucket for ", urlloc)
						repoStats.Inc(common.Stored)
					}
					// TODO  Is here where to add an entry to the KV store
					db.Update(func(tx *bolt.Tx) error {
						b := tx.Bucket([]byte(sourceName))
						err := b.Put([]byte(urlloc), []byte(sha))
						if err != nil {
							log.Error("Error writing to bolt ", err)
						}
						return nil
					})
				} else {
					log.WithFields(log.Fields{"url": urlloc, "issue": "Empty JSON-LD document found "}).Info("Empty JSON-LD document found. Continuing.")
					repologger.WithFields(log.Fields{"url": urlloc, "issue": "Empty JSON-LD document found "}).Error(err)
					// TODO  Is here where to add an entry to the KV store
					db.Update(func(tx *bolt.Tx) error {
						b := tx.Bucket([]byte(sourceName))
						err := b.Put([]byte(urlloc), []byte(fmt.Sprintf("NULL: %s", urlloc))) // no JOSN-LD found at this URL
						if err != nil {
							log.Error("Error writing to bolt ", err)
						}
						return nil
					})
				}
			}

			bar.Add(1)                                          // bar.Incr()
			log.Trace("#", i, "thread for", urlloc)             // print an message containing the index (won't keep order)
			time.Sleep(time.Duration(delay) * time.Millisecond) // sleep a bit if directed to by the provider

			lwg.Done()

			<-semaphoreChan // clear a spot in the semaphore channel for the next indexing event
		}(i, sourceName)

	}
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
