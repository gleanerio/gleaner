package acquire

import (
	"github.com/PuerkitoBio/goquery"
	"fmt"
	"github.com/gleanerio/gleaner/internal/common"
	configTypes "github.com/gleanerio/gleaner/internal/config"
	"github.com/minio/minio-go/v7"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	bolt "go.etcd.io/bbolt"
	"net/http"
	"sync"
	"time"
)

/* Acquire JSON-LD from API endpoints */
const APIType = "api"

// Read the config and get API endpoint template strings
func RetrieveAPIEndpoints(v1 *viper.Viper) ([]configTypes.Sources, error) {
	var apiSources []configTypes.Sources

	// Get our API sources
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
func RetrieveAPIData(apiSources []configTypes.Sources, mc *minio.Client, db *bolt.DB, runStats *common.RunStats, v1 *viper.Viper) {
	wg := sync.WaitGroup{}

	for _, source := range apiSources {
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

func getAPISource(v1 *viper.Viper, mc *minio.Client, source configTypes.Sources, wg *sync.WaitGroup, db *bolt.DB, repologger *log.Logger, repoStats *common.RepoStats) {
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

	responseStatusChan := make(chan int, tc) // a blocking channel to keep concurrency under control
	lwg := sync.WaitGroup{}

	defer func() {
		lwg.Wait()
		wg.Done()
		close(responseStatusChan)
	}()

	// Loop through our paged API template until
	// we get an error; i is the page number in this case
	status := http.StatusOK // start off with an OK default
	i := 0
	for status == http.StatusOK {
		lwg.Add(1)
		urlloc := fmt.Sprintf(source.URL, i)

		go func(i int, sourceName string) {
			repologger.Trace("Indexing", urlloc)
			log.Debug("Indexing ", urlloc)

			req, err := http.NewRequest("GET", urlloc, nil)
			if err != nil {
				log.Error(i, err, urlloc)
			}
			req.Header.Set("User-Agent", EarthCubeAgent)
			req.Header.Set("Accept", "application/ld+json, text/html")

			response, err := client.Do(req)

			if err != nil {
				log.Error("#", i, " error on ", urlloc, err) // print an message containing the index (won't keep order)
				repologger.WithFields(log.Fields{"url": urlloc}).Error(err)
				lwg.Done()
				responseStatusChan <- http.StatusBadRequest
				return
			}

			if response.StatusCode != http.StatusOK {
				log.Error("#", i, " response status ", response.StatusCode, " from ", urlloc)
				repologger.WithFields(log.Fields{"url": urlloc}).Error(response.StatusCode)
				lwg.Done()
				responseStatusChan <- response.StatusCode
				return
			}

			defer response.Body.Close()
			log.Trace("Response status ", response.StatusCode, " from ", urlloc)
			responseStatusChan <- response.StatusCode

			jsonlds, err := findJSONInResponse(response)

			if err != nil {
				log.Error("#", i, " error on ", urlloc, err) // print an message containing the index (won't keep order)
				repoStats.Inc(common.Issues)
				lwg.Done() // tell the wait group that we be done
				responseStatusChan <- http.StatusBadRequest
				return
			}

			// even if no JSON-LD packages found, record the event of checking this URL
			if len(jsonlds) < 1 {
				log.WithFields(log.Fields{"url": urlloc, "contentType": "No JSON-LD found']"}).Info("No JSON-LD found at ", urlloc)
				repologger.WithFields(log.Fields{"url": urlloc, "contentType": "No JSON-LD found']"}).Error() // this needs to go into the issues file

				db.Update(func(tx *bolt.Tx) error {
					b := tx.Bucket([]byte(sourceName))
					err := b.Put([]byte(urlloc), []byte(fmt.Sprintf("NILL: %s", urlloc))) // no JSON-LD found at this URL
					return err
				})

			} else {
				log.WithFields(log.Fields{"url": urlloc, "issue": "Indexed"}).Trace("Indexed ", urlloc)
				repologger.WithFields(log.Fields{"url": urlloc, "issue": "Indexed"}).Trace()
				repoStats.Inc(common.Summoned)
			}

			UploadWrapper(v1, mc, bucketName, sourceName, urlloc, db, repologger, repoStats, jsonlds)

			log.Trace("#", i, "thread for", urlloc)             // print an message containing the index (won't keep order)
			time.Sleep(time.Duration(delay) * time.Millisecond) // sleep a bit if directed to by the provider

			lwg.Done()
		}(i, source.Name)
		status = <- responseStatusChan
		i++
	}
}
