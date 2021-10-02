package acquire

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gleanerio/gleaner/internal/common"
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
	for k := range m {
		log.Printf("Headless chrome call to: %s", k)

		for i := range m[k] {
			err := PageRender(v1, mc, 60*time.Second, m[k][i], k) // TODO make delay configurable
			if err != nil {
				log.Printf("%s :: %s", m[k][i], err)
			}
		}
	}
}

func PageRender(v1 *viper.Viper, mc *minio.Client, timeout time.Duration, url, k string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	mcfg := v1.GetStringMapString("summoner")

	// Use the DevTools HTTP/JSON API to manage targets (e.g. pages, webworkers).
	devt := devtool.New(mcfg["headless"])
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
 * and then tries to find a JSON-LD element on the page, and get its contents.
 *  If it doesn't find one, it will retry three times, with a wait in between. You can see that it
 *  ultimately calls reject() with no arguments if it can't find anything, and that is because
 * I cannot figure out how to get the cdp Runtime to distinguish between a resolved and a rejected
 *  promise - so in this case, we simply do not index a document, and fail silently.
 **/
	expression := `
		function getMetadata() {
			return new Promise((resolve, reject) => {
				const element = document.querySelector('script[type="application/ld+json"]');
				if(element && element.innerText) {
					const metadata = element.innerText;
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
						reject("");
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

	// todo: what are the data types that will always be in this json? we
	// could create a struct out of them if we want to.
	var jsonld string
	if err = json.Unmarshal(eval.Result.Value, &jsonld); err != nil {
		log.Println(err)
		return (err)
	}

	if jsonld != "" { // traps out the root domain...   should do this different
		// get sha1 of the JSONLD..  it's a nice ID
		fmt.Printf("%s JSON-LD: %s\n\n", url, jsonld)

		h := sha1.New()
		h.Write([]byte(jsonld))
		bs := h.Sum(nil)
		bss := hex.EncodeToString(bs[:])

		sha, err := common.GetNormSHA(jsonld, v1) // Moved to the normalized sha value
		if err != nil {
			log.Println(err)
		}

		objectName := fmt.Sprintf("summoned/%s/%s.jsonld", k, sha)

		contentType := "application/ld+json"
		b := bytes.NewBufferString(jsonld)

		usermeta := make(map[string]string) // what do I want to know?
		usermeta["url"] = url
		usermeta["sha1"] = bss
		bucketName := "gleaner"

		err = StoreProv(v1, mc, k, sha, url)
		if err != nil {
			log.Println(err)
		}

		// Upload the  file with FPutObject
		n, err := mc.PutObject(context.Background(), bucketName, objectName, b, int64(b.Len()), minio.PutObjectOptions{ContentType: contentType, UserMetadata: usermeta})
		if err != nil {
			log.Printf("Error uploading %s: %s", objectName, err)
		}
		log.Printf("Uploaded Bucket:%s File:%s Size %d \n", bucketName, objectName, n.Size)
	} else {
		log.Println("No JSON-LD found at", url)
	}

	return nil
}
