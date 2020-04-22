package acquire

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"earthcube.org/Project418/gleaner/internal/common"
	"earthcube.org/Project418/gleaner/pkg/summoner/sitemaps"
	"github.com/PuerkitoBio/goquery"
	"github.com/gosuri/uiprogress"
	"github.com/minio/minio-go"
	"github.com/spf13/viper"
)

// ResRetrieve is a function to pull down the data graphs at resources
func ResRetrieve(v1 *viper.Viper, mc *minio.Client, m map[string]sitemaps.URLSet) {
	uiprogress.Start()
	wg := sync.WaitGroup{}

	for k := range m {
		log.Printf("Queuing URLs for %s \n", k)
		go getDomain(v1, mc, m, k, &wg)
	}

	time.Sleep(2 * time.Second) // ?? why is this here?
	wg.Wait()
	uiprogress.Stop()
}

func getDomain(v1 *viper.Viper, mc *minio.Client, m map[string]sitemaps.URLSet, k string, wg *sync.WaitGroup) {
	semaphoreChan := make(chan struct{}, 10) // a blocking channel to keep concurrency under control
	defer close(semaphoreChan)
	lwg := sync.WaitGroup{}

	wg.Add(1)       // wg from the calling function
	defer wg.Done() // tell the wait group that we be done

	count := len(m[k].URL)
	bar := uiprogress.AddBar(count).PrependElapsed().AppendCompleted()
	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return rightPad2Len(k, " ", 25)
	})
	bar.Fill = '-'
	bar.Head = '>'
	bar.Empty = ' '

	// if count < 1 {
	// 	log.Printf("No resources found for %s \n", k)
	// 	return // should maked this return an error
	// }

	var (
		buf    bytes.Buffer
		logger = log.New(&buf, "logger: ", log.Lshortfile)
	)

	// we actually go get the URLs now
	for i := range m[k].URL {
		lwg.Add(1)
		urlloc := m[k].URL[i].Loc

		// TODO / WARNING for large site we can exhaust memory with just the creation of the
		// go routines. 1 million =~ 4 GB  So we need to control how many routines we
		// make too..  reference https://github.com/mr51m0n/gorc (but look for someting in the core
		// library too)

		go func(i int, k string) {
			semaphoreChan <- struct{}{}

			var client http.Client // why do I make this here..  can I use 1 client?  move up in the loop
			req, err := http.NewRequest("GET", urlloc, nil)
			if err != nil {
				logger.Printf("#%d error on %s : %s  ", i, urlloc, err) // print an message containing the index (won't keep order)
			}

			req.Header.Set("User-Agent", "EarthCube_DataBot/1.0")

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

			var jsonld string

			// TODO  this check should be for application/ld+json, not octet stream
			if err == nil && !contains(resp.Header["Content-Type"], "application/ld+json") {
				doc.Find("script").Each(func(i int, s *goquery.Selection) {
					val, _ := s.Attr("type")
					if val == "application/ld+json" {
						action, err := isValid(v1, s.Text())
						if err != nil {
							logger.Printf("ERROR: URL: %s Action: %s  Error: %s", urlloc, action, err)
						}
						jsonld = s.Text()
					}
				})
			}

			if err == nil && contains(resp.Header["Content-Type"], "application/ld+json") {
				jsonld = doc.Text()
			}

			if jsonld != "" { // traps out the root domain...   should do this different
				sha, err := common.GetNormSHA(jsonld, v1) // Moved to the normalized sha value
				if err != nil {
					logger.Printf("ERROR: URL: %s Action: Getting normalized sha  Error: %s\n", urlloc, err)
				}
				objectName := fmt.Sprintf("%s/%s.jsonld", k, sha)
				contentType := "application/ld+json"
				b := bytes.NewBufferString(jsonld)

				usermeta := make(map[string]string) // what do I want to know?
				usermeta["url"] = urlloc
				usermeta["sha1"] = sha
				bucketName := "gleaner-summoned" //   fmt.Sprintf("gleaner-summoned/%s", k) // old was just k

				// Upload the file with FPutObject
				_, err = mc.PutObject(bucketName, objectName, b, int64(b.Len()), minio.PutObjectOptions{ContentType: contentType, UserMetadata: usermeta})
				if err != nil {
					logger.Printf("%s", objectName)
					logger.Fatalln(err) // Fatal?   seriously?    I guess this is the object write, so the run is likely a bust at this point, but this seems a bit much still.
				}
				// logger.Printf("#%d Uploaded Bucket:%s File:%s Size %d\n", i, bucketName, objectName, n)
			}

			bar.Incr()

			logger.Printf("#%d thread for %s ", i, urlloc) // print an message containing the index (won't keep order)
			lwg.Done()                                     // tell the wait group that we be done

			<-semaphoreChan // clear a spot in the semaphore channel
		}(i, k)

	}

	lwg.Wait()

	// TODO write this to minio in the run ID bucket
	// return the logger buffer or write to a mutex locked bytes buffer
	f, err := os.Create(fmt.Sprintf("./%s.log", k))
	if err != nil {
		log.Println("Error writing a file")
	}

	w := bufio.NewWriter(f)
	_, err = w.WriteString(buf.String())
	if err != nil {
		log.Println("Error writing a file")
	}
	w.Flush()

}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func rightPad2Len(s string, padStr string, overallLen int) string {
	var padCountInt int
	padCountInt = 1 + ((overallLen - len(padStr)) / len(padStr))
	var retStr = s + strings.Repeat(padStr, padCountInt)
	return retStr[:overallLen]
}

func isValid(v1 *viper.Viper, jsonld string) (string, error) {
	proc, options := common.JLDProc(v1)

	var myInterface interface{}
	action := ""

	err := json.Unmarshal([]byte(jsonld), &myInterface)
	if err != nil {
		action = "json.Unmarshal call"
		return action, err
	}

	_, err = proc.ToRDF(myInterface, options) // returns triples but toss them, just validating
	if err != nil {                           // it's wasted cycles.. but if just doing a summon, needs to be done here
		action = "JSON-LD to RDF call"
		return action, err
	}

	return action, err
}
