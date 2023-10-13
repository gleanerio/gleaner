package acquire

import (
	"context"
	"github.com/chromedp/chromedp"
	"github.com/gleanerio/gleaner/internal/common"
	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/target"
	"github.com/mafredri/cdp/rpcc"
	"github.com/mafredri/cdp/session"
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
)

const EarthCubeAgent = "EarthCube_DataBot/1.0"
const JSONContentType = "application/ld+json"

// ResRetrieve is a function to pull down the data graphs at resources
func ResRetrieve(v1 *viper.Viper, mc *minio.Client, m map[string][]string, runStats *common.RunStats) {
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
		go getDomain(v1, mc, urls, domain, &wg, repologger, r)
	}

	wg.Wait()
}

func getConfig(v1 *viper.Viper, sourceName string) (string, int, int64, int, string, string, error, bool) {
	bucketName, err := configTypes.GetBucketName(v1)
	if err != nil {
		return bucketName, 0, 0, 0, configTypes.AccceptContentType, "", err, false
	}

	var mcfg configTypes.Summoner
	mcfg, err = configTypes.ReadSummmonerConfig(v1.Sub("summoner"))

	if err != nil {
		return bucketName, 0, 0, 0, configTypes.AccceptContentType, "", err, false
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
	acceptContent := source.AcceptContentType
	jsonProfile := source.JsonProfile
	hw := source.HeadlessWait
	headless := source.Headless
	if err != nil {
		return bucketName, tc, delay, hw, acceptContent, jsonProfile, err, false
	}

	if source.Delay != 0 && source.Delay > delay {
		delay = source.Delay
		tc = 1
		log.Info("Crawl delay set to ", delay, " for ", sourceName)
	}
	log.Info("Thread count ", tc, " delay ", delay)

	return bucketName, tc, delay, hw, acceptContent, jsonProfile, nil, headless
}

func getDomain(v1 *viper.Viper, mc *minio.Client, urls []string, sourceName string,
	wg *sync.WaitGroup, repologger *log.Logger, repoStats *common.RepoStats) {

	var timeout = 60 * time.Second
	var retries = 3
	var totalTimeout = timeout * time.Duration(retries+1)

	bucketName, tc, delay, headlessWait, acceptContent, jsonProfile, err, headless := getConfig(v1, sourceName)
	if err != nil {
		// trying to read a source, so let's not kill everything with a panic/fatal
		log.Error("Error reading config file ", err)
		repologger.Error("Error reading config file ", err)
	}

	var client http.Client

	// stuff to setup headless sessions
	if headlessWait < 0 {
		log.Info("Headless wait on a headless configured to less that zero. Setting to 0")
		headlessWait = 0 // if someone screws up the config, be good
	}

	if totalTimeout < time.Duration(headlessWait)*time.Second {
		timeout = time.Duration(headlessWait) * time.Second
	}
	/// if you cancel here, then everything after first times out
	//ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Duration(retries))
	//ctx, cancel := context.WithTimeout(context.TODO(), timeout*time.Duration(retries))
	//defer cancel()
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// read config file
	mcfg, err := configTypes.ReadSummmonerConfig(v1.Sub("summoner"))

	// Use the DevTools HTTP/JSON API to manage targets (e.g. pages, webworkers).
	//devt := devtool.New(mcfg["headless"])
	devt := devtool.New(mcfg.Headless)

	pt, err := devt.Get(ctx, devtool.Page)
	if err != nil {
		pt, err = devt.Create(ctx)
		if err != nil {
			log.WithFields(log.Fields{"issue": "Not REPO FAULT. Devtools... Is Headless Container running?"}).Error(err)
			repologger.WithFields(log.Fields{}).Error("Not REPO FAULT. Devtools... Is Headless Container running?")
			repoStats.Inc(common.HeadlessError)
			return
		}
	}

	// Initiate a new RPC connection to the Chrome DevTools Protocol target.
	conn, err := rpcc.DialContext(ctx, pt.WebSocketDebuggerURL)
	if err != nil {
		log.WithFields(log.Fields{"issue": "Not REPO FAULT. Devtools... Is Headless Container running?"}).Error(err)
		repologger.WithFields(log.Fields{}).Error("Not REPO FAULT. Devtools... Is Headless Container running?")
		repoStats.Inc(common.HeadlessError)
		return
	}
	defer conn.Close()
	sessionclient := cdp.NewClient(conn)
	// headless_agent := emulation.NewSetUserAgentOverrideArgs(EarthCubeAgent)
	// https://www.zenrows.com/blog/chromedp#user-agent-in-chromedp
	//https://pkg.go.dev/github.com/mafredri/cdp/devtool#WithClient
	m, err := session.NewManager(sessionclient)
	if err != nil {
		// Handle error.
	}
	defer m.Close()

	// session

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

			if headless {
				log.WithFields(log.Fields{"url": urlloc, "issue": "running headless"}).Trace("Headless ", urlloc)
				args := target.NewCreateTargetArgs("")
				//args.SetNewWindow(true)
				newPage, err := sessionclient.Target.CreateTarget(ctx,
					args)
				if err != nil {
					log.WithFields(log.Fields{"url": urlloc, "issue": "Not REPO FAULT. NewCreateTargetArgs... Is Headless Container running?"}).Error(err)
					repologger.WithFields(log.Fields{"url": urlloc}).Error("Not REPO FAULT. NewCreateTargetArgs... Is Headless Container running?")
					repoStats.Inc(common.HeadlessError)
					lwg.Done()
					<-semaphoreChan
					return
				}
				closeArgs := target.NewCloseTargetArgs(newPage.TargetID)
				defer func(Target cdp.Target, ctx context.Context, args *target.CloseTargetArgs) {
					log.Infof("Close Target Defer targetID: %s  url: %s ", newPage.TargetID, urlloc)
					_, err := Target.CloseTarget(ctx, args)
					if err != nil {
						log.WithFields(log.Fields{"url": urlloc, "issue": "error closing target"}).Error("PageRenderAndUpload ", urlloc, "::", err)

					}
				}(sessionclient.Target, ctx, closeArgs)
				// newPageConn uses the underlying conn without establishing a new
				// websocket connection.
				//newPageConn, err := m.Dial(ctx, newPage.TargetID)
				//if err != nil {
				//	log.WithFields(log.Fields{"url": urlloc, "issue": "Not REPO FAULT. newPageConn... Is Headless Container running?"}).Error(err)
				//	repologger.WithFields(log.Fields{"url": urlloc}).Error("Not REPO FAULT. newPageConn... Is Headless Container running?")
				//	repoStats.Inc(common.HeadlessError)
				//	lwg.Done()
				//	<-semaphoreChan
				//	return
				//}
				////defer func(newPageConn *rpcc.Conn) {
				////	log.Info("NewPageConn defer")
				////	err := newPageConn.Close()
				////	if err != nil {
				////		log.WithFields(log.Fields{"url": urlloc, "issue": "error clocing connection"}).Error("PageRenderAndUpload ", urlloc, "::", err)
				////
				////	}
				////}(newPageConn)
				//
				//c := cdp.NewClient(newPageConn)
				err = PageRenderAndUpload(v1, mc, timeout, urlloc, sourceName, repologger, repoStats, m, newPage.TargetID) // TODO make delay configurable

				if err != nil {
					log.WithFields(log.Fields{"url": urlloc, "issue": "converting json ld"}).Error("PageRenderAndUpload ", urlloc, " ::", err)
					repologger.WithFields(log.Fields{"url": urlloc, "issue": "converting json ld"}).Error(err)
					//err = newPageConn.Close()
					//if err != nil {
					//	log.WithFields(log.Fields{"url": urlloc, "issue": "error closing connection"}).Error("PageRenderAndUpload ", urlloc, " ::", err)
					//
					//}
					//closeTargetResp, err := sessionclient.Target.CloseTarget(ctx, closeArgs)
					//log.Info(closeTargetResp)
					//if err != nil {
					//	log.WithFields(log.Fields{"url": urlloc, "issue": "error closing target"}).Error("PageRenderAndUpload ", urlloc, " ::", err)
					//
					//}
					lwg.Done()
					<-semaphoreChan
					return
				}
				//err = newPageConn.Close()
				//if err != nil {
				//	log.WithFields(log.Fields{"url": urlloc, "issue": "error closing connection"}).Error("PageRenderAndUpload ", urlloc, " ::", err)
				//
				//}
				//closeTargetResp, err := sessionclient.Target.CloseTarget(ctx, closeArgs)
				//log.Info(closeTargetResp)
				//if err != nil {
				//	log.WithFields(log.Fields{"url": urlloc, "issue": "error closing target"}).Error("PageRenderAndUpload ", urlloc, " ::", err)
				//
				//}
			} else {
				req, err := http.NewRequest("GET", urlloc, nil)
				if err != nil {
					log.Error(i, err, urlloc)
				}
				req.Header.Set("User-Agent", EarthCubeAgent)
				req.Header.Set("Accept", acceptContent)

				resp, err := client.Do(req)
				if err != nil {
					log.Error("#", i, " error on ", urlloc, err) // print an message containing the index (won't keep order)
					repologger.WithFields(log.Fields{"url": urlloc}).Error(err)
					lwg.Done() // tell the wait group that we be done
					<-semaphoreChan
					return
				}
				defer resp.Body.Close()

				// if there is an error, then don't try again.
				if (resp.StatusCode >= 400) && (resp.StatusCode < 600) {
					switch resp.StatusCode {
					case 403:
						log.Error("#", i, " not authorized ", urlloc, err) // print an message containing the index (won't keep order)
						repologger.WithFields(log.Fields{"url": urlloc}).Error(err)
						repoStats.Inc(common.NotAuthorized)
					case 404:
						log.Error("#", i, " bad url ", urlloc, err) // print an message containing the index (won't keep order)
						repologger.WithFields(log.Fields{"url": urlloc}).Error(err)
						repoStats.Inc(common.BadUrl)
					case 500:
						log.Error("#", i, " server arror ", urlloc, err) // print an message containing the index (won't keep order)
						repologger.WithFields(log.Fields{"url": urlloc}).Error(err)
						repoStats.Inc(common.RepoServerError)
					default:
						log.Error("#", i, " generic arror ", urlloc, err) // print an message containing the index (won't keep order)
						repologger.WithFields(log.Fields{"url": urlloc}).Error(err)
						repoStats.Inc(common.GenericIssue)
					}
					lwg.Done() // tell the wait group that we be done
					<-semaphoreChan
					return
				}

				jsonlds, err := FindJSONInResponse(v1, urlloc, jsonProfile, repologger, resp)
				// there was an issue with sitemaps... but now this code
				//if contains(contentTypeHeader, JSONContentType) || contains(contentTypeHeader, "application/json") {
				//
				//	b, err := io.ReadAll(resp.Body)
				//	// b, err := ioutil.ReadAll(resp.Body)  Go.1.15 and earlier
				//	if err != nil {
				//		log.Error("#", i, " error on ", urlloc, err) // print an message containing the index (won't keep order)
				//		repoStats.Inc(common.Issues)
				//		lwg.Done() // tell the wait group that we be done
				//		<-semaphoreChan
				//		return
				//	}
				//	jsonlds = []string{string(b)}
				//} else {
				//	var err error
				//	jsonlds, err = FindJSONInResponse(v1, urlloc, jsonProfile, repologger, resp)
				//	if err != nil {
				//		log.Error("#", i, " error on ", urlloc, err) // print an message containing the index (won't keep order)
				//		repoStats.Inc(common.Issues)
				//		lwg.Done() // tell the wait group that we be done
				//		<-semaphoreChan
				//		return
				//	}
				//}
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
					if headlessWait >= 0 {
						log.WithFields(log.Fields{"url": urlloc, "contentType": "Direct access failed, trying headless']"}).Info("Direct access failed, trying headless for ", urlloc)
						repologger.WithFields(log.Fields{"url": urlloc, "contentType": "Direct access failed, trying headless']"}).Error() // this needs to go into the issues file
						args := target.NewCreateTargetArgs("")
						//args.SetNewWindow(true)
						newPage, err := sessionclient.Target.CreateTarget(ctx,
							args)
						if err != nil {
							log.WithFields(log.Fields{"url": urlloc, "issue": "Not REPO FAULT. NewCreateTargetArgs... Is Headless Container running?"}).Error(err)
							repologger.WithFields(log.Fields{"url": urlloc}).Error("Not REPO FAULT. NewCreateTargetArgs... Is Headless Container running?")
							repoStats.Inc(common.HeadlessError)
							lwg.Done()
							<-semaphoreChan
							return
						}
						closeArgs := target.NewCloseTargetArgs(newPage.TargetID)
						defer func(Target cdp.Target, ctx context.Context, args *target.CloseTargetArgs) {
							log.Info("Close Target Defer")
							_, err := Target.CloseTarget(ctx, args)
							if err != nil {
								log.WithFields(log.Fields{"url": urlloc, "issue": "error closing target"}).Error("PageRenderAndUpload ", urlloc, "::", err)

							}
						}(sessionclient.Target, ctx, closeArgs)
						// newPageConn uses the underlying conn without establishing a new
						// websocket connection.
						//newPageConn, err := m.Dial(ctx, newPage.TargetID)
						//if err != nil {
						//	log.WithFields(log.Fields{"url": urlloc, "issue": "Not REPO FAULT. newPageConn... Is Headless Container running?"}).Error(err)
						//	repologger.WithFields(log.Fields{"url": urlloc}).Error("Not REPO FAULT. newPageConn... Is Headless Container running?")
						//	repoStats.Inc(common.HeadlessError)
						//	lwg.Done()
						//	<-semaphoreChan
						//	return
						//}
						//defer func(newPageConn *rpcc.Conn) {
						//	log.Info("NewPageConn defer")
						//	err := newPageConn.Close()
						//	if err != nil {
						//		log.WithFields(log.Fields{"url": urlloc, "issue": "error clocing connection"}).Error("PageRenderAndUpload ", urlloc, "::", err)
						//
						//	}
						//}(newPageConn)
						//
						//c := cdp.NewClient(newPageConn)
						err = PageRenderAndUpload(v1, mc, 60*time.Second, urlloc, sourceName, repologger, repoStats, m, newPage.TargetID) // TODO make delay configurable
						if err != nil {
							log.WithFields(log.Fields{"url": urlloc, "issue": "converting json ld"}).Error("PageRenderAndUpload ", urlloc, "::", err)
							repologger.WithFields(log.Fields{"url": urlloc, "issue": "converting json ld"}).Error(err)
						}
					}

				} else {
					log.WithFields(log.Fields{"url": urlloc, "issue": "Direct access worked"}).Trace("Direct access worked for ", urlloc)
					repologger.WithFields(log.Fields{"url": urlloc, "issue": "Direct access worked"}).Trace()
					repoStats.Inc(common.Summoned)
				}

				UploadWrapper(v1, mc, bucketName, sourceName, urlloc, repologger, repoStats, jsonlds)
			} // else headless
			bar.Add(1)                                          // bar.Incr()
			log.Trace("#", i, "thread for", urlloc)             // print an message containing the index (won't keep order)
			time.Sleep(time.Duration(delay) * time.Millisecond) // sleep a bit if directed to by the provider

			lwg.Done()

			<-semaphoreChan // clear a spot in the semaphore channel for the next indexing event
		}(i, sourceName)

	}
	common.RunRepoStatsOutput(repoStats, sourceName)
}

func FindJSONInResponse(v1 *viper.Viper, urlloc string, jsonProfile string, repologger *log.Logger, response *http.Response) ([]string, error) {
	doc, err := goquery.NewDocumentFromResponse(response)
	if err != nil {
		return nil, err
	}

	contentTypeHeader := response.Header["Content-Type"]
	var jsonlds []string

	// if the URL is sending back JSON-LD correctly as application/ld+json
	// this should not be here IMHO, but need to support people not setting proper header value
	// The URL is sending back JSON-LD but incorrectly sending as application/json
	// would like to add contains(contentTypeHeader, jsonProfile)
	// but empty profile strings matching all
	if contains(contentTypeHeader, JSONContentType) || contains(contentTypeHeader, "application/json") || fileExtensionIsJson(urlloc) {
		logFields := log.Fields{"url": urlloc, "contentType": "json or ld_json"}
		repologger.WithFields(logFields).Debug()
		log.WithFields(logFields).Debug(urlloc, " as ", contentTypeHeader)
		resp_text := doc.Text()
		jsonlds, err = addToJsonListIfValid(v1, jsonlds, resp_text)
		if err != nil {
			log.WithFields(logFields).Error("Error processing json response from ", urlloc, err)
			repologger.WithFields(logFields).Error(err)
		}
		// look in the HTML response for <script type=application/ld+json> ^
	} else {
		//doc.Find("script[type='application/ld+json']").Each(func(i int, s *goquery.Selection) {
		//Please note that Cascadia's selectors do not necessarily match all supported selectors of jQuery (Sizzle).  https://github.com/andybalholm/cascadia
		doc.Find("script[type^='application/ld+json']").Each(func(i int, s *goquery.Selection) {
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

func UploadWrapper(v1 *viper.Viper, mc *minio.Client, bucketName string, sourceName string, urlloc string, repologger *log.Logger, repoStats *common.RepoStats, jsonlds []string) {
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

		} else {
			logFields := log.Fields{"url": urlloc, "issue": "Empty JSON-LD document found "}
			log.WithFields(logFields).Info("Empty JSON-LD document found. Continuing.")
			repologger.WithFields(logFields).Error("Empty JSON-LD document found. Continuing.")

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
