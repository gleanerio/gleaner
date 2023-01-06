package acquire

import (
	"fmt"
	"github.com/gleanerio/gleaner/internal/common"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	configTypes "github.com/gleanerio/gleaner/internal/config"

	"github.com/PuerkitoBio/goquery"
	"github.com/minio/minio-go/v7"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/viper"
	bolt "go.etcd.io/bbolt"
)

const EarthCubeAgent = "EarthCube_DataBot/1.0"
const JSONContentType = "application/ld+json"

// ResRetrieve is a function to pull down the data graphs at resources
func ResRetrieve(v1 *viper.Viper, mc *minio.Client, m map[string][]string, db *bolt.DB, runStats *common.RunStats) {
	wg := sync.WaitGroup{}

	// Why do I pass the wg pointer?   Just make a new one
	// for each domain in getDomain and us this one here with a semaphore
	// to control the loop?
	for domain, urls := range m {
		r := runStats.Add(domain)
		r.Set(common.Count, len(urls))
		r.Set(common.HttpError, 0)
		r.Set(common.Issues, 0)
		r.Set(common.Summoned, 0)
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

			jsonlds, err := FindJSONInResponse(v1, urlloc, repologger, resp)

			if err != nil {
				log.Error("#", i, " error on ", urlloc, err) // print an message containing the index (won't keep order)
				repoStats.Inc(common.Issues)
				lwg.Done() // tell the wait group that we be done
				<-semaphoreChan
				return
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

			UploadWrapper(v1, mc, bucketName, sourceName, urlloc, db, repologger, repoStats, jsonlds)

			bar.Add(1)                                          // bar.Incr()
			log.Trace("#", i, "thread for", urlloc)             // print an message containing the index (won't keep order)
			time.Sleep(time.Duration(delay) * time.Millisecond) // sleep a bit if directed to by the provider

			lwg.Done()

			<-semaphoreChan // clear a spot in the semaphore channel for the next indexing event
		}(i, sourceName)

	}
}

func FindJSONInResponse(v1 *viper.Viper, urlloc string, repologger *log.Logger, response *http.Response) ([]string, error) {
	doc, err := goquery.NewDocumentFromResponse(response)
	if err != nil {
		return nil, err
	}

	contentTypeHeader := response.Header["Content-Type"]
	var jsonlds []string

	// if the URL is sending back JSON-LD correctly as application/ld+json
	// this should not be here IMHO, but need to support people not setting proper header value
	// The URL is sending back JSON-LD but incorrectly sending as application/json
	if contains(contentTypeHeader, JSONContentType) || contains(contentTypeHeader, "application/json") || fileExtensionIsJson(urlloc) {
		logFields := log.Fields{"url": urlloc, "contentType": "json or ld_json"}
		repologger.WithFields(logFields).Debug()
		log.WithFields(logFields).Debug(urlloc, " as ", contentTypeHeader)

		jsonlds, err = addToJsonListIfValid(v1, jsonlds, doc.Text())
		if err != nil {
			log.WithFields(logFields).Error("Error processing json response from ", urlloc, err)
			repologger.WithFields(logFields).Error(err)
		}
		// look in the HTML response for <script type=application/ld+json>
	} else {
		doc.Find("script[type='application/ld+json']").Each(func(i int, s *goquery.Selection) {
			jsonlds, err = addToJsonListIfValid(v1, jsonlds, s.Text())
			logFields := log.Fields{"url": urlloc, "contentType": "script[type='application/ld+json']"}
			repologger.WithFields(logFields).Info()
			if err != nil {
				log.WithFields(logFields).Error("Error processing script tag in ", urlloc, err)
				repologger.WithFields(logFields).Error(err)
			}
		})
	}

	return jsonlds, nil
}

func UploadWrapper(v1 *viper.Viper, mc *minio.Client, bucketName string, sourceName string, urlloc string, db *bolt.DB, repologger *log.Logger, repoStats *common.RepoStats, jsonlds []string) {
	for i, jsonld := range jsonlds {
		if jsonld != "" { // traps out the root domain...   should do this different
			logFields := log.Fields{"url": urlloc, "issue": "Uploading"}
			log.WithFields(logFields).Trace("#", i, "Uploading ")
			repologger.WithFields(logFields).Trace()
			sha, err := Upload(v1, mc, bucketName, sourceName, urlloc, jsonld)
			if err != nil {
				logFields = log.Fields{"url": urlloc, "sha": sha, "issue": "Error uploading jsonld to object store"}
				log.WithFields(logFields).Error("Error uploading jsonld to object store: ", urlloc, err)
				repologger.WithFields(logFields).Error(err)
				repoStats.Inc(common.StoreError)
			} else {
				logFields = log.Fields{"url": urlloc, "sha": sha, "issue": "Uploaded to object store"}
				repologger.WithFields(logFields).Trace(err)
				log.WithFields(logFields).Info("Successfully put ", sha, " in summoned bucket for ", urlloc)
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
			logFields := log.Fields{"url": urlloc, "issue": "Empty JSON-LD document found "}
			log.WithFields(logFields).Info("Empty JSON-LD document found. Continuing.")
			repologger.WithFields(logFields).Error("Empty JSON-LD document found. Continuing.")
			// TODO  Is here where to add an entry to the KV store
			db.Update(func(tx *bolt.Tx) error {
				b := tx.Bucket([]byte(sourceName))
				err := b.Put([]byte(urlloc), []byte(fmt.Sprintf("NULL: %s", urlloc))) // no JSON-LD found at this URL
				if err != nil {
					log.Error("Error writing to bolt ", err)
				}
				return nil
			})
		}
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
