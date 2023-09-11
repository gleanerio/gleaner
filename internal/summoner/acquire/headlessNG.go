package acquire

import (
	"context"
	"encoding/json"
	"fmt"
	target2 "github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"github.com/gleanerio/gleaner/internal/common"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/target"
	"github.com/mafredri/cdp/rpcc"
	"github.com/mafredri/cdp/session"
	log "github.com/sirupsen/logrus"
	"time"

	configTypes "github.com/gleanerio/gleaner/internal/config"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/protocol/runtime"
	minio "github.com/minio/minio-go/v7"
	"github.com/spf13/viper"
	"github.com/valyala/fasttemplate"
)

// HeadlessNG gets schema.org entries in sites that put the JSON-LD in dynamically with JS.
// It uses a chrome headless instance (which MUST BE RUNNING).
// TODO..  trap out error where headless is NOT running
func HeadlessNG(v1 *viper.Viper, mc *minio.Client, m map[string][]string, runStats *common.RunStats) {
	// NOTE   this function compares to ResRetrieve in acquire.go.  They both approach things
	// in same ways due to hwo we deal with threading (opportunities).   We don't queue up domains
	// multiple times since we are dealing with our local resource now in the form of the headless tooling.

	//var (
	//	buf    bytes.Buffer
	//	logger = log.New(&buf, "logger: ", log.Lshortfile)
	//)
	var timeout = 60 * time.Second
	var retries = 3
	var totalTimeout = timeout * time.Duration(retries+1)

	for k := range m {
		r := runStats.Add(k)
		r.Set(common.Count, len(m[k]))
		log.Trace("Headless chrome call to:", k)
		repologger, err := common.LogIssues(v1, k)
		if err != nil {
			log.Error("Headless Error creating a logger for a repository", err)

		} else {
			repologger.Info("Headless chrome call to ", k)
		}
		_, _, _, headlessWait, _, _, err, _ := getConfig(v1, k)
		if err != nil {
			// trying to read a source, so let's not kill everything with a panic/fatal
			log.Error("Error reading config file ", err)
			repologger.Error("Error reading config file ", err)
		}
		// stuff to setup headless sessions
		if headlessWait < 0 {
			log.Info("Headless wait on a headless configured to less that zero. Setting to 0")
			headlessWait = 0 // if someone screws up the config, be good
		}

		if totalTimeout < time.Duration(headlessWait)*time.Second {
			timeout = time.Duration(headlessWait) * time.Second
		}
		//ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Duration(retries))
		//ctx, cancel := context.WithTimeout(context.TODO(), timeout*time.Duration(retries))
		//defer cancel()
		ctx, cancel := chromedp.NewContext(context.TODO())
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

				return
			}
		}

		// Initiate a new RPC connection to the Chrome DevTools Protocol target.
		conn, err := rpcc.DialContext(ctx, pt.WebSocketDebuggerURL)
		if err != nil {
			log.WithFields(log.Fields{"issue": "Not REPO FAULT. Devtools... Is Headless Container running?"}).Error(err)
			repologger.WithFields(log.Fields{}).Error("Not REPO FAULT. Devtools... Is Headless Container running?")

			return
		}
		defer conn.Close()
		sessionclient := cdp.NewClient(conn)
		manager, err := session.NewManager(sessionclient)
		if err != nil {
			// Handle error.
		}
		defer manager.Close()

		// session
		for i := range m[k] {
			args := target.NewCreateTargetArgs("")
			//args.SetNewWindow(true)
			newPage, err := sessionclient.Target.CreateTarget(ctx,
				args)
			if err != nil {
				log.WithFields(log.Fields{"url": m[k][i], "issue": "Not REPO FAULT. NewCreateTargetArgs... Is Headless Container running?"}).Error(err)
				repologger.WithFields(log.Fields{"url": m[k][i]}).Error("Not REPO FAULT. NewCreateTargetArgs... Is Headless Container running?")

				return
			}
			closeArgs := target.NewCloseTargetArgs(newPage.TargetID)
			defer func(Target cdp.Target, ctx context.Context, args *target.CloseTargetArgs) {
				log.Info("Close Target Defer")
				_, err := Target.CloseTarget(ctx, args)
				if err != nil {
					log.WithFields(log.Fields{"url": m[k][i], "issue": "error closing target"}).Error("PageRenderAndUpload ", m[k][i], " ::", err)

				}
			}(sessionclient.Target, ctx, closeArgs)
			err = PageRenderAndUpload(v1, mc, timeout, m[k][i], k, repologger, r, manager, newPage.TargetID) // TODO make delay configurable
			if err != nil {
				log.Error(m[k][i], "::", err)
			}
		}
		common.RunRepoStatsOutput(r, k)

	}

}

//// ThreadedHeadlessNG does not work.. ;)
//func ThreadedHeadlessNG(v1 *viper.Viper, mc *minio.Client, m map[string][]string, db *bolt.DB) {
//	wg := sync.WaitGroup{}
//
//	//var (
//	//	buf    bytes.Buffer
//	//	logger = log.New(&buf, "logger: ", log.Lshortfile)
//	//)
//
//	for k := range m {
//		log.Trace("Headless chrome call to:", k)
//
//		// for i := range m[k] {
//		go doCall(v1, mc, 60*time.Second, m, k, &wg, db) // TODO make delay configurable
//		// 	log.Printf("%s :: %s", m[k][i], err)
//		// }
//	}
//
//	time.Sleep(2 * time.Second) // ?? why is this here?
//	wg.Wait()
//
//}
//
//func doCall(v1 *viper.Viper, mc *minio.Client, timeout time.Duration, m map[string][]string, k string, wg *sync.WaitGroup, db *bolt.DB) {
//
//	db.Update(func(tx *bolt.Tx) error {
//		_, err := tx.CreateBucket([]byte(k))
//		if err != nil {
//			return fmt.Errorf("create bucket: %s", err)
//		}
//		return nil
//	})
//
//	tc, err := Threadcount(v1)
//	if err != nil {
//		log.Error(err)
//	}
//	dt, err := Delayrequest(v1)
//	if err != nil {
//		log.Error(err)
//	}
//
//	if dt > 0 {
//		tc = 1 // If the domain requests a delay between request, drop to single threaded and honor delay
//	}
//
//	log.Info("Thread count", tc, "delay", dt)
//
//	// thread vars
//	semaphoreChan := make(chan struct{}, tc) // a blocking channel to keep concurrency under control
//	defer close(semaphoreChan)
//	lwg := sync.WaitGroup{}
//
//	wg.Add(1)       // wg from the calling function
//	defer wg.Done() // tell the wait group that we be done
//
//	for i := range m[k] {
//		lwg.Add(1)
//		urlloc := m[k][i]
//
//		go func(i int, k string) {
//			//thread management
//			semaphoreChan <- struct{}{}
//
//			err := PageRenderAndUpload(v1, mc, 60*time.Second, m[k][i], k, db) // TODO make delay configurable
//			if err != nil {
//				log.Error(m[k][i], "::", err)
//			}
//
//			// thread management
//			log.Debug("#", i, "thread for", urlloc)          // print an message containing the index (won't keep order)
//			time.Sleep(time.Duration(dt) * time.Millisecond) // sleep a bit if directed to by the provider
//
//			lwg.Done()
//			<-semaphoreChan // clear a spot in the semaphore channel for the next indexing event
//
//		}(i, k)
//	}
//
//	lwg.Wait()
//
//}

func PageRenderAndUpload(v1 *viper.Viper, mc *minio.Client, timeout time.Duration, url, k string, repologger *log.Logger, repoStats *common.RepoStats, m *session.Manager, targetID target.ID) error {
	repologger.WithFields(log.Fields{"url": url}).Trace("PageRenderAndUpload")
	// page render handles this
	//ctx, cancel := context.WithTimeout(context.Background(), timeout)
	//defer cancel()

	// read config file
	//miniocfg := v1.GetStringMapString("minio")
	//bucketName := miniocfg["bucket"] //   get the top level bucket for all of gleaner operations from config file
	bucketName, err := configTypes.GetBucketName(v1)

	//mcfg := v1.GetStringMapString("summoner")
	//mcfg, err := configTypes.ReadSummmonerConfig(v1.Sub("summoner"))

	jsonlds, err := PageRender(v1, timeout, url, k, repologger, repoStats, m, targetID)

	if err == nil { // from page render. If there are no errros, upload.
		if len(jsonlds) > 1 {
			log.WithFields(log.Fields{"url": url, "issue": "Multiple JSON"}).Info("Error uploading jsonld to object store:", url)
			repologger.WithFields(log.Fields{"url": url, "issue": "Multiple JSON"}).Info()
		}
		for i, jsonld := range jsonlds {
			sha, err2 := Upload(v1, mc, bucketName, k, url, jsonld)
			if err2 != nil {
				log.WithFields(log.Fields{"url": url, "sha": sha, "jsonld#": i, "issue": "Error uploading jsonld to object store"}).Error("Error uploading jsonld to object store:", url, err2, sha)
				repologger.WithFields(log.Fields{"url": url, "sha": sha, "jsonld#": i, "issue": "Error uploading jsonld to object store"}).Error(err2)
				repoStats.Inc(common.StoreError)
			} else {
				log.WithFields(log.Fields{"url": url, "sha": sha, "jsonld#": i, "issue": "Uploaded JSONLD to object store"}).Info("Uploaded JSONLD to object store:", url, sha)
				repologger.WithFields(log.Fields{"url": url, "sha": sha, "jsonld#": i, "issue": "Uploaded JSONLD to object store"}).Info()
				repoStats.Inc(common.Stored)
			}
		}
	}
	return err
}

func PageRender(v1 *viper.Viper, timeout time.Duration, url, k string, repologger *log.Logger, repoStats *common.RepoStats, m *session.Manager, targetID target.ID) ([]string, error) {

	repologger.WithFields(log.Fields{"url": url}).Trace("PageRender")
	retries := 3
	var totalTimeout = timeout * time.Duration(retries+1) //add some time for cleanup.
	sources, err := configTypes.GetSources(v1)
	source, err := configTypes.GetSourceByName(sources, k)
	headlessWait := source.HeadlessWait
	response := []string{}
	if headlessWait < 0 {
		log.Info("Headless wait on a headless configured to less that zero. Setting to 0")
		headlessWait = 0 // if someone screws up the config, be good
	}

	if timeout*time.Duration(retries) < time.Duration(headlessWait)*time.Second {
		timeout = time.Duration(headlessWait) * time.Second
	}

	// context for rpcc.Conn with timeout.
	//ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Duration(retries))
	//ctxrRcc, cancel := context.WithTimeout(context.TODO(), timeout*time.Duration(retries))
	//defer cancel()
	ctxPrcc, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	newPageConn, err := m.Dial(ctxPrcc, targetID)
	log.Infof("headless context targetID: %s  url: %s", targetID, url)
	if err != nil {
		log.WithFields(log.Fields{"url": url, "issue": "Not REPO FAULT. newPageConn... Is Headless Container running?"}).Error(err)
		repologger.WithFields(log.Fields{"url": url}).Error("Not REPO FAULT. newPageConn... Is Headless Container running?")
		repoStats.Inc(common.HeadlessError)

		return nil, err
	}
	//defer func(newPageConn *rpcc.Conn) {
	//	log.Info("NewPageConn defer")
	//	err := newPageConn.Close()
	//	if err != nil {
	//		log.WithFields(log.Fields{"url": urlloc, "issue": "error clocing connection"}).Error("PageRenderAndUpload ", urlloc, "::", err)
	//
	//	}
	//}(newPageConn)

	// context for page rendering. Separate from the rpcc.Conn
	ctx, cancel := context.WithTimeout(context.TODO(), totalTimeout)
	defer cancel()
	ctx, cancel = chromedp.NewContext(ctx, chromedp.WithTargetID(target2.ID(targetID)))
	defer cancel()

	client := cdp.NewClient(newPageConn)

	//ctx, cancel := context.WithTimeout(context.TODO(), timeout*time.Duration(retries))

	//
	//// read config file
	//mcfg, err := configTypes.ReadSummmonerConfig(v1.Sub("summoner"))
	//
	//// Use the DevTools HTTP/JSON API to manage targets (e.g. pages, webworkers).
	////devt := devtool.New(mcfg["headless"])
	//devt := devtool.New(mcfg.Headless)
	//
	//pt, err := devt.Get(ctx, devtool.Page)
	//if err != nil {
	//	pt, err = devt.Create(ctx)
	//	if err != nil {
	//		log.WithFields(log.Fields{"url": url, "issue": "Not REPO FAULT. Devtools... Is Headless Container running?"}).Error(err)
	//		repologger.WithFields(log.Fields{"url": url}).Error("Not REPO FAULT. Devtools... Is Headless Container running?")
	//		repoStats.Inc(common.HeadlessError)
	//		return response, err
	//	}
	//}
	//
	//// Initiate a new RPC connection to the Chrome DevTools Protocol target.
	//conn, err := rpcc.DialContext(ctx, pt.WebSocketDebuggerURL)
	//if err != nil {
	//	log.WithFields(log.Fields{"url": url, "issue": "Not REPO FAULT. Devtools... Is Headless Container running?"}).Error(err)
	//	repologger.WithFields(log.Fields{"url": url}).Error("Not REPO FAULT. Devtools... Is Headless Container running?")
	//	repoStats.Inc(common.HeadlessError)
	//	return response, err
	//}
	//defer conn.Close() // Leaving connections open will leak memory.
	//
	//// attempt to use session. failed.
	////sessionclient := cdp.NewClient(conn) // conn created via rpcc.Dial.
	////m, err := session.NewManager(sessionclient)
	////if err != nil {
	////	// Handle error.
	////}
	////defer m.Close()
	////
	////newPage, err := sessionclient.Target.CreateTarget(ctx,
	////	target.NewCreateTargetArgs("about:blank"))
	////if err != nil {
	////	log.WithFields(log.Fields{"url": url, "issue": "Not REPO FAULT. NewCreateTargetArgs... Is Headless Container running?"}).Error(err)
	////	repologger.WithFields(log.Fields{"url": url}).Error("Not REPO FAULT. NewCreateTargetArgs... Is Headless Container running?")
	////	repoStats.Inc(common.HeadlessError)
	////	return response, err
	////}
	////
	////// newPageConn uses the underlying conn without establishing a new
	////// websocket connection.
	////newPageConn, err := m.Dial(ctx, newPage.TargetID)
	////if err != nil {
	////	log.WithFields(log.Fields{"url": url, "issue": "Not REPO FAULT. newPageConn... Is Headless Container running?"}).Error(err)
	////	repologger.WithFields(log.Fields{"url": url}).Error("Not REPO FAULT. newPageConn... Is Headless Container running?")
	////	repoStats.Inc(common.HeadlessError)
	////	return response, err
	////}
	////defer newPageConn.Close()
	////
	////c := cdp.NewClient(newPageConn)
	//
	//c := cdp.NewClient(conn)
	c := client
	// Listen to Page events so we can receive DomContentEventFired, which
	// is what tells us when the page is done loading
	err = c.Page.Enable(ctx)
	if err != nil {
		log.WithFields(log.Fields{"url": url, "issue": "Not REPO FAULT. Devtools... Is Headless Container running?"}).Error(err)
		repologger.WithFields(log.Fields{"url": url}).Error("Not REPO FAULT. Devtools... Is Headless Container running?")
		repoStats.Inc(common.HeadlessError)
		return response, err
	}

	// Open a DOMContentEventFired client to buffer this event.
	domContent, err := c.Page.DOMContentEventFired(ctx)
	if err != nil {
		log.WithFields(log.Fields{"url": url, "issue": "Not REPO FAULT. Devtools... Is Headless Container running?"}).Error(err)
		repologger.WithFields(log.Fields{"url": url}).Error("Not REPO FAULT. Devtools... Is Headless Container running?")
		repoStats.Inc(common.HeadlessError)
		return response, err
	}
	defer domContent.Close()

	// Open a LoadEventFired client to buffer this event.
	loadEventFired, err := c.Page.LoadEventFired(ctx)
	if err != nil {
		log.WithFields(log.Fields{"url": url, "issue": "Not REPO FAULT. Devtools... Is Headless Container running?"}).Error(err)
		repologger.WithFields(log.Fields{"url": url}).Error("Not REPO FAULT. Devtools... Is Headless Container running?")
		repoStats.Inc(common.HeadlessError)
		return response, err
	}
	defer loadEventFired.Close()

	// Create the Navigate arguments with the optional Referrer field set.
	navArgs := page.NewNavigateArgs(url)
	nav, err := c.Page.Navigate(ctx, navArgs)

	if err != nil {
		log.WithFields(log.Fields{"url": url, "issue": "Navigate To Headless"}).Error(err)
		repologger.WithFields(log.Fields{"url": url, "issue": "Navigate To Headless"}).Error(err)
		repoStats.Inc(common.HeadlessError)
		return response, err
	}

	if loadEventReply, err := loadEventFired.Recv(); err != nil {
		log.Trace(loadEventReply)
		log.Errorf(" loadEventFired error  original targetID: %s  url: %s", targetID, url)
		log.Errorf("loadEventFired  error context    TargetID: %s  url: %s", ctx.Value("targetId"), url)
		log.WithFields(log.Fields{"url": url, "issue": "Headless Load Event Error"}).Error(err)
		repologger.WithFields(log.Fields{"url": url, "issue": "Headless Load Event Error"}).Error(err)
		repoStats.Inc(common.HeadlessError)
		return response, err
	}
	loadEventFired.Close()

	// Wait until we have a DOMContentEventFired event.
	if contentReply, err := domContent.Recv(); err != nil {
		log.Trace(contentReply)
		log.Errorf(" domContent.Recv error  original targetID: %s  url: %s", targetID, url)
		log.Errorf("domContent.Recv  error context    TargetID: %s  url: %s", ctx.Value("targetId"), url)
		log.WithFields(log.Fields{"url": url, "issue": "Dom Error"}).Error(err)
		repologger.WithFields(log.Fields{"url": url, "issue": "Dom Error"}).Error(err)
		repoStats.Inc(common.HeadlessError)
		return response, err
	}

	log.WithFields(log.Fields{"url": url, "issue": "Navigate Complete"}).Debug(nav.FrameID, "for", url)
	repologger.WithFields(log.Fields{"url": url, "issue": "Navigate Complete"}).Trace()
	/**
	 * This JavaScript expression will be run in Headless Chrome. It waits for 1000 milliseconds,
	 * and then tries to find all of the JSON-LD elements on the page, and get their contents.
	 *  If it doesn't find one, it will retry three times, with a wait in between. You can see that it
	 *  ultimately calls reject() with no arguments if it can't find anything, and that is because
	 * I cannot figure out how to get the cdp Runtime to distinguish between a resolved and a rejected
	 *  promise - so in this case, we simply do not index a document, and fail silently.
	 **/
	expressionTmpl := `
		function getMetadata() {
			return new Promise((resolve, reject) => {
				const elements = document.querySelectorAll('script[type^="application/ld+json"]');
				let metadata = [];
				elements.forEach(function(element) {
					if(element && element.innerText) {
						metadata.push(element.innerText);
					}
				})
				if(metadata.length) {
					resolve(metadata);
				}
				else {
					reject("No JSON-LD present after {{timeout}} milliseconds.");
				}
			});
		}

		function retry(fn, retriesLeft = {{retries}}, interval = {{timeout}}) {
			return new Promise((resolve, reject) => {
				fn()
					.then(resolve)
					.catch((error) => {
						if (retriesLeft === 0) {
						reject(null);
						return;
					}

					setTimeout(() => {
						retry(fn, retriesLeft - 1, interval).then(resolve).catch(reject);
					}, interval);
				});
			});
		}
		function sleep(ms) {
		  return new Promise(resolve => setTimeout(resolve, ms));
		}

        sleep( {{headlesswait}} ).then( () => { return retry(getMetadata) } );
	
	`
	tmpl := fasttemplate.New(expressionTmpl, "{{", "}}")
	expression := tmpl.ExecuteString(map[string]interface{}{
		"timeout":      fmt.Sprintf("%d", timeout.Milliseconds()),
		"headlesswait": fmt.Sprintf("%d", headlessWait*1000),
		"retries":      "3",
	})
	log.Trace(expression)
	evalArgs := runtime.NewEvaluateArgs(expression).SetAwaitPromise(true).SetReturnByValue(true)
	eval, err := c.Runtime.Evaluate(ctx, evalArgs)
	if err != nil {
		log.WithFields(log.Fields{"url": url, "issue": "Headless Evaluate"}).Error(err)
		repologger.WithFields(log.Fields{"url": url, "issue": "Headless Evaluate"}).Error(err)
		repoStats.Inc(common.Issues)
		return response, err
	}

	// Rejecting that promise just sends null as its value,
	// so we need to stop if we got that.
	if eval.Result.Value == nil {
		repologger.WithFields(log.Fields{"url": url, "issue": "Headless Nil Result"}).Trace()
		repoStats.Inc(common.EmptyDoc)
		return response, nil
	}

	// todo: what are the data types that will always be in this json? we
	// could create a struct out of them if we want to.
	var jsonlds []string
	if err = json.Unmarshal(eval.Result.Value, &jsonlds); err != nil {
		log.WithFields(log.Fields{"url": url, "issue": "Json Unmarshal"}).Error(err)
		repologger.WithFields(log.Fields{"url": url, "issue": "Json Unmarshal"}).Error(err)
		repoStats.Inc(common.Issues)
		return response, err
	}

	if len(jsonlds) > 1 {
		repologger.WithFields(log.Fields{"url": url, "issue": "Multiple JSON"}).Debug(err)
	}
	for _, jsonld := range jsonlds {
		// just grab them all maybe we fix later in upload process.
		if jsonld != "" {
			response = append(response, jsonld)
		}

		//valid, err := isValid(v1, jsonld)
		//if err != nil {
		//	// there could be one bad jsonld, and one good. We want to process the jsonld
		//	// so, do not set an err
		//	log.WithFields(log.Fields{"url": url, "issue": "invalid JSON"}).Error("error checking for valid json :", err)
		//	repologger.WithFields(log.Fields{"url": url, "issue": "invalid JSON"}).Error(err)
		//	repoStats.Inc(common.Issues)
		//} else if valid && jsonld != "" { // traps out the root domain...   should do this different
		//	response = append(response, jsonld)
		//	err = nil
		//	// need to just return a list
		//
		//} else {
		//	// there could be one bad jsonld, and one good. We want to process the jsonld
		//	// so, do not set an err
		//	log.Info("Empty JSON-LD document found. Continuing.", url)
		//	repologger.WithFields(log.Fields{"url": url, "issue": "Empty JSON-LD document found"}).Error()
		//	repoStats.Inc(common.EmptyDoc)
		//	// TODO  Is here where to add an entry to the KV store
		//	//err = db.Update(func(tx *bolt.Tx) error {
		//	//	b := tx.Bucket([]byte(k))
		//	//	err := b.Put([]byte(url), []byte("NULL")) // no JOSN-LD found at this URL
		//	//	if err != nil {
		//	//		log.Error("Error writing to bolt", err)
		//	//	}
		//	//	return nil
		//	//})
		//}
	}

	return response, err
}
