package acquire

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"github.com/gleanerio/gleaner/internal/common"
	configTypes "github.com/gleanerio/gleaner/internal/config"
	"github.com/gleanerio/gleaner/internal/millers/graph"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	"github.com/minio/minio-go/v7"
	bolt "go.etcd.io/bbolt"
)

/* Acquire JSON-LD from API endpoints */
const APIType = "api"

// Read the config and get API endpoint template strings
func RetrieveAPIEndpoints(v1 *viper.Viper) ([]string, error) {
	var apiSources []string

	// Get our API sources
	mcfg, err := configTypes.ReadSummmonerConfig(v1.Sub("summoner"))
	sources, err := configTypes.GetSources(v1)
	if err != nil {
		log.Error("Error getting sources to summon: ", err)
		return apiSources, err
	}

	apiSources = configTypes.GetActiveSourceByType(sources, APIType)
	return apiSources, err
}

// given a paged API url template, iterate through the pages until we get
// all the results we want.
func RetrieveAPIData(apiSources map[string][]string, mc *minio.Client, db *bolt.DB, runStats *common.RunStats) {
	wg := sync.WaitGroup{}

	for source := range apiSources {
		r := runStats.Add(source.Name)
		r.Set(common.HttpError, 0)
		r.Set(common.Issues, 0)
		r.Set(common.Summoned, 0)
		log.Info("Queuing API calls for ", source.Name)

		repologger, err := common.LogIssues(v1, source.Name)
		if err != nil {
			log.Error("Error creating a logger for a repository", err)
		} else {
			repologger.Info("Queuing API calls for ", source.Name)
		}
		wg.Add(1)
		go getAPISource(v1, mc, source, &wg, db, repologger, r)
	}

	wg.Wait()
}

func getAPISource(v1 *viper.Viper, mc *minio.Client, apiSource map[string][]string, wg *sync.WaitGroup, db *bolt.DB, repologger *log.Logger, repoStats *common.RepoStats) {
	// make the bucket (if it doesn't exist)
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(source.Name))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	bucketName, tc, delay, err := getConfig(v1, source.Name)
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

	// Loop through our paged API template until
	// we get an error; i is the page number in this case
	var response http.Response
	i := 0
	for response.StatusCode && response.StatusOK {
		urlloc := fmt.Sprintf(source.url, page)

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

			log.Trace("#", i, "thread for", urlloc)             // print an message containing the index (won't keep order)
			time.Sleep(time.Duration(delay) * time.Millisecond) // sleep a bit if directed to by the provider

			lwg.Done()

			<-semaphoreChan // clear a spot in the semaphore channel for the next indexing event
		}(i, sourceName)
		i++
	}
}
