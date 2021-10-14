package acquire

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/minio/minio-go/v7"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/viper"
)

// ResRetrieve is a function to pull down the data graphs at resources
func ResRetrieve(v1 *viper.Viper, mc *minio.Client, m map[string][]string) {
	wg := sync.WaitGroup{}

	// Why do I pass the wg pointer?   Just make a new one
	// for each domain in getDomain and us this one here with a semaphore
	// to control the loop?
	for k := range m {
		// log.Printf("Queuing URLs for %s \n", k)
		go getDomain(v1, mc, m, k, &wg)
	}

	time.Sleep(2 * time.Second) // ?? why is this here?
	wg.Wait()
}

func getDomain(v1 *viper.Viper, mc *minio.Client, m map[string][]string, k string, wg *sync.WaitGroup) {

	// read config file
	miniocfg := v1.GetStringMapString("minio")
	bucketName := miniocfg["bucket"] //   get the top level bucket for all of gleaner operations from config file

	mcfg := v1.GetStringMapString("summoner")
	tc, err := strconv.ParseInt(mcfg["threads"], 10, 64)
	if err != nil {
		log.Println(err)
		log.Panic("Could not convert threads from config file to an int")
	}

	delay := mcfg["delay"]
	var dt int64
	if delay != "" {
		//log.Printf("Delay set to: %s milliseconds", delay)
		dt, err = strconv.ParseInt(delay, 10, 64)
		if err != nil {
			log.Println(err)
			log.Panic("Could not convert delay from config file to a value")
		}
		// set threads to 1
		//log.Println("Delay is not 0, threads set to 1")
		tc = 1
	} else {
		dt = 0
	}

	semaphoreChan := make(chan struct{}, tc) // a blocking channel to keep concurrency under control
	defer close(semaphoreChan)
	lwg := sync.WaitGroup{}

	wg.Add(1)       // wg from the calling function
	defer wg.Done() // tell the wait group that we be done

	count := len(m[k])
	// OLD bar
	// bar := uiprogress.AddBar(count).PrependElapsed().AppendCompleted()
	// bar.PrependFunc(func(b *uiprogress.Bar) string {
	// 	return rightPad2Len(k, " ", 15)
	// })
	// bar.Fill = '-'
	// bar.Head = '>'
	// bar.Empty = ' '

	bar := progressbar.Default(int64(count))

	// if count < 1 {
	// 	log.Printf("No resources found for %s \n", k)
	// 	return // should maked this return an error
	// }

	var (
		buf    bytes.Buffer
		logger = log.New(&buf, "logger: ", log.Lshortfile)
	)

	// var client http.Client

	// we actually go get the URLs now
	for i := range m[k] {
		lwg.Add(1)
		urlloc := m[k][i]

		// TODO / WARNING for large site we can exhaust memory with just the creation of the
		// go routines. 1 million =~ 4 GB  So we need to control how many routines we
		// make too..  reference https://github.com/mr51m0n/gorc (but look for someting in the core
		// library too)

		// log.Println(urlloc)

		go func(i int, k string) {
			semaphoreChan <- struct{}{}

			// log.Println(urlloc)

			urlloc = strings.ReplaceAll(urlloc, " ", "")
			urlloc = strings.ReplaceAll(urlloc, "\n", "")

			var client http.Client // why do I make this here..  can I use 1 client?  move up in the loop
			req, err := http.NewRequest("GET", urlloc, nil)
			if err != nil {
				log.Println(err)
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

			var jsonlds []string
			var contentTypeHeader = resp.Header["Content-Type"]

			if (
			// The URL is sending back JSON-LD correctly as application/ld+json
			contains(contentTypeHeader, "application/ld+json") ||

			// this should not be here IMHO, but need to support people not setting proper header value
			// The URL is sending back JSON-LD but incorrectly sending as application/json
			contains(contentTypeHeader, "application/json")){
				jsonlds, err = addToJsonListIfValid(v1, jsonlds, doc.Text())
				if err != nil {
					logger.Printf("Error processing json response from %s: %s", urlloc, err)
				}
			// look in the HTML page for <script type=application/ld+json
			} else {
				doc.Find("script").Each(func(i int, s *goquery.Selection) {
					val, _ := s.Attr("type")
					if val == "application/ld+json" {
						jsonlds, err = addToJsonListIfValid(v1, jsonlds, s.Text())
						if err != nil {
							logger.Printf("Error processing script tag in %s: %s", urlloc, err)
						}
					}
				})
			}

			for _, jsonld := range jsonlds {
				if jsonld != "" { // traps out the root domain...   should do this different
					logger.Printf("#%d Uploading", i)
					Upload(v1, mc, logger, bucketName, k, urlloc, jsonld)
				} else {
					logger.Printf("Empty JSON-LD document found. Continuing.")
				}
			}

			bar.Add(1) // bar.Incr()

			logger.Printf("#%d thread for %s ", i, urlloc) // print an message containing the index (won't keep order)
			lwg.Done()

			time.Sleep(time.Duration(dt) * time.Millisecond) // tell the wait group that we be done

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
		// if a == str {
		if strings.Contains(a, str) {
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
