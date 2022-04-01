package acquire

import (
	"context"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"

	configTypes "github.com/gleanerio/gleaner/internal/config"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/protocol/runtime"
	"github.com/mafredri/cdp/rpcc"
	minio "github.com/minio/minio-go/v7"
	"github.com/spf13/viper"
)

// HeadlessNG gets schema.org entries in sites that put the JSON-LD in dynamically with JS.
// It uses a chrome headless instance (which MUST BE RUNNING).
// TODO..  trap out error where headless is NOT running
func HeadlessNG(v1 *viper.Viper, mc *minio.Client, m map[string][]string) {
	// NOTE   this function compares to ResRetrieve in acquire.go.  They both approach things
	// in same ways due to hwo we deal with threading (opportunities).   We don't queue up domains
	// multiple times since we are dealing with our local resource now in the form of the headless tooling.

	//var (
	//	buf    bytes.Buffer
	//	logger = log.New(&buf, "logger: ", log.Lshortfile)
	//)

	for k := range m {
		log.Printf("Headless chrome call to: %s", k)
		for i := range m[k] {
			err := PageRender(v1, mc, 60*time.Second, m[k][i], k) // TODO make delay configurable
			if err != nil {
				log.Printf("%s :: %s", m[k][i], fmt.Sprintf("%s", err))
			}
		}

	}

}

// ThreadedHeadlessNG does not work.. ;)
func ThreadedHeadlessNG(v1 *viper.Viper, mc *minio.Client, m map[string][]string) {
	wg := sync.WaitGroup{}

	//var (
	//	buf    bytes.Buffer
	//	logger = log.New(&buf, "logger: ", log.Lshortfile)
	//)

	for k := range m {
		log.Printf("Headless chrome call to: %s", k)

		// for i := range m[k] {
		go doCall(v1, mc, 60*time.Second, m, k, &wg) // TODO make delay configurable
		// 	log.Printf("%s :: %s", m[k][i], err)
		// }
	}

	time.Sleep(2 * time.Second) // ?? why is this here?
	wg.Wait()

}

func doCall(v1 *viper.Viper, mc *minio.Client, timeout time.Duration, m map[string][]string, k string, wg *sync.WaitGroup) {

	tc, err := Threadcount(v1)
	if err != nil {
		log.Println(err)
	}
	dt, err := Delayrequest(v1)
	if err != nil {
		log.Println(err)
	}

	if dt > 0 {
		tc = 1 // If the domain requests a delay between request, drop to single threaded and honor delay
	}

	log.Printf("Thread count %d delay %d\n", tc, dt)

	// thread vars
	semaphoreChan := make(chan struct{}, tc) // a blocking channel to keep concurrency under control
	defer close(semaphoreChan)
	lwg := sync.WaitGroup{}

	wg.Add(1)       // wg from the calling function
	defer wg.Done() // tell the wait group that we be done

	for i := range m[k] {
		lwg.Add(1)
		urlloc := m[k][i]

		go func(i int, k string) {
			//thread management
			semaphoreChan <- struct{}{}

			err := PageRender(v1, mc, 60*time.Second, m[k][i], k) // TODO make delay configurable
			if err != nil {
				log.Printf("%s :: %s", m[k][i], err)
			}

			// thread management
			log.Printf("#%d thread for %s ", i, urlloc)      // print an message containing the index (won't keep order)
			time.Sleep(time.Duration(dt) * time.Millisecond) // sleep a bit if directed to by the provider

			lwg.Done()
			<-semaphoreChan // clear a spot in the semaphore channel for the next indexing event

		}(i, k)
	}

	lwg.Wait()

}

func PageRender(v1 *viper.Viper, mc *minio.Client, timeout time.Duration, url, k string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// read config file
	//miniocfg := v1.GetStringMapString("minio")
	//bucketName := miniocfg["bucket"] //   get the top level bucket for all of gleaner operations from config file
	bucketName, err := configTypes.GetBucketName(v1)

	//mcfg := v1.GetStringMapString("summoner")
	mcfg, err := configTypes.ReadSummmonerConfig(v1.Sub("summoner"))

	// Use the DevTools HTTP/JSON API to manage targets (e.g. pages, webworkers).
	//devt := devtool.New(mcfg["headless"])
	devt := devtool.New(mcfg.Headless)
	pt, err := devt.Get(ctx, devtool.Page)
	if err != nil {
		pt, err = devt.Create(ctx)
		if err != nil {
			log.Print(err)
			return err
		}
	}

	// Initiate a new RPC connection to the Chrome DevTools Protocol target.
	conn, err := rpcc.DialContext(ctx, pt.WebSocketDebuggerURL)
	if err != nil {
		log.Print(err)
		return err
	}
	defer conn.Close() // Leaving connections open will leak memory.

	c := cdp.NewClient(conn)

	// Listen to Page events so we can receive DomContentEventFired, which
	// is what tells us when the page is done loading
	err = c.Page.Enable(ctx)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Open a DOMContentEventFired client to buffer this event.
	domContent, err := c.Page.DOMContentEventFired(ctx)
	if err != nil {
		log.Print(err)
		return err
	}
	defer domContent.Close()

	// Create the Navigate arguments with the optional Referrer field set.
	navArgs := page.NewNavigateArgs(url)
	nav, err := c.Page.Navigate(ctx, navArgs)
	if err != nil {
		log.Print(err)
		return err
	}

	// Wait until we have a DOMContentEventFired event.
	if _, err = domContent.Recv(); err != nil {
		log.Print(err)
		return err
	}

	log.Printf("%s for %s\n", nav.FrameID, url)

	/**
	 * This JavaScript expression will be run in Headless Chrome. It waits for 1000 milliseconds,
	 * and then tries to find all of the JSON-LD elements on the page, and get their contents.
	 *  If it doesn't find one, it will retry three times, with a wait in between. You can see that it
	 *  ultimately calls reject() with no arguments if it can't find anything, and that is because
	 * I cannot figure out how to get the cdp Runtime to distinguish between a resolved and a rejected
	 *  promise - so in this case, we simply do not index a document, and fail silently.
	 **/
	expression := `
		function getMetadata() {
			return new Promise((resolve, reject) => {
				const elements = document.querySelectorAll('script[type="application/ld+json"]');
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
					reject("No JSON-LD present after 1 second.");
				}
			});
		}

		function retry(fn, retriesLeft = 3, interval = 1000) {
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

		retry(getMetadata);
	`

	evalArgs := runtime.NewEvaluateArgs(expression).SetAwaitPromise(true).SetReturnByValue(true)
	eval, err := c.Runtime.Evaluate(ctx, evalArgs)
	if err != nil {
		log.Println(err)
		return (err)
	}

	// Rejecting that promise just sends null as its value,
	// so we need to stop if we got that.
	if eval.Result.Value == nil {
		return nil
	}

	// todo: what are the data types that will always be in this json? we
	// could create a struct out of them if we want to.
	var jsonlds []string
	if err = json.Unmarshal(eval.Result.Value, &jsonlds); err != nil {
		log.Println(err)
		return (err)
	}

	for _, jsonld := range jsonlds {
		valid, err := isValid(v1, jsonld)
		if err != nil {
			log.Printf("error checking for valid json: %s", err)
		} else if valid && jsonld != "" { // traps out the root domain...   should do this different
			sha, err := Upload(v1, mc, bucketName, k, url, jsonld)
			if err != nil {
				log.Printf("Error uploading jsonld to object store: %s: %s: %s", url, err, sha)
			}

		} else {
			log.Println("Empty JSON-LD document found. Continuing.", url)

		}
	}

	return err
}
