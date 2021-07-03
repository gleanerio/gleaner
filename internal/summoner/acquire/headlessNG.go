package acquire

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/earthcubearchitecture-project418/gleaner/internal/common"
	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/network"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/protocol/runtime"
	"github.com/mafredri/cdp/rpcc"
	minio "github.com/minio/minio-go/v7"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
)

// DocumentInfo contains information about the document.
type DocumentInfo struct {
	Title string `json:"title"`
}

// Cookie represents a browser cookie.
type Cookie struct {
	URL   string `json:"url"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

// HeadlessNG gets schema.org entries in sites that put the JSON-LD in dynamically with JS.
// It uses a chrome headless instance (which MUST BE RUNNING).
// TODO..  trap out error where headless is NOT running
func HeadlessNG(v1 *viper.Viper, mc *minio.Client, m map[string][]string) {
	for k := range m {
		log.Printf("Headless chrome call to: %s", k)

		for i := range m[k] {
			err := PageRender(v1, mc, 30*time.Second, m[k][i], k) // TODO make delay configurable
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

	// Open a DOMContentEventFired client to buffer this event.
	domContent, err := c.Page.DOMContentEventFired(ctx)
	if err != nil {
		return err

	}
	defer domContent.Close()

	// Give enough capacity to avoid blocking any event listeners
	//abort := make(chan error, 2)

	// Watch the abort channel.
	//go func() {
	//select {
	//case <-ctx.Done():
	//case err := <-abort:
	//fmt.Printf("aborted: %s\n", err.Error())
	//cancel()
	//}
	//}()

	// Setup event handlers early because domain events can be sent as
	// soon as Enable is called on the domain.
	//if err = abortOnErrors(ctx, c, abort); err != nil {
	//fmt.Println(err)
	//return err
	//}

	if err = runBatch(
		// Enable all the domain events that we're interested in.
		func() error { return c.DOM.Enable(ctx) },
		func() error { return c.Network.Enable(ctx, nil) },
		func() error { return c.Page.Enable(ctx) },
		func() error { return c.Runtime.Enable(ctx) },

		// func() error { return setCookies(ctx, c.Network, Cookies...) },
	); err != nil {
		fmt.Println(err)
		return err
	}

	// Open a DOMContentEventFired client to buffer this event.
	domContent, err = c.Page.DOMContentEventFired(ctx)
	if err != nil {
		log.Print(err)
		return err
	}
	defer domContent.Close()

	// Enable events on the Page domain, it's often preferrable to create
	// event clients before enabling events so that we don't miss any.
	// if err = c.Page.Enable(ctx); err != nil {
	// 	log.Print(err)
	// 	return err
	// }

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

	// Parse information from the document by evaluating JavaScript.
	// const title = document.getElementById('geocodes').innerText;
	// const title = document.querySelector('script[id="geocodes"]').innerText;

	/**
	const title = document.querySelector('script[type="application/ld+json"]').innerText;
	const title = document.querySelector('#jsonld').innerText;
	const title = document.querySelector('#geocodes').innerText;
	const title = document.querySelector('script[id="geocodes"]').innerText;
	**/

	expression := `
		new Promise((resolve, reject) => {
			setTimeout(() => {
				const title = document.querySelector('script[type="application/ld+json"]').innerText;
				resolve({title});
			}, 1000);
		});
	`

	// expression := `
	// 	new Promise((resolve, reject) => {
	// 		setTimeout(() => {
	// 			const title = document.querySelectorAll('script[type="application/ld+json"]')[0].innerText;
	// 			resolve({title});
	// 		}, 1000);
	// 	});
	// `

	evalArgs := runtime.NewEvaluateArgs(expression).SetAwaitPromise(true).SetReturnByValue(true)
	eval, err := c.Runtime.Evaluate(ctx, evalArgs)
	if err != nil {
		log.Println(err)
		return (err)
	}

	var info DocumentInfo
	if err = json.Unmarshal(eval.Result.Value, &info); err != nil {
		log.Println(err)
		return (err)
	}

	jsonld := info.Title
	fmt.Printf("%s JSON-LD: %s\n\n", url, jsonld)

	if info.Title != "" { // traps out the root domain...   should do this different
		// get sha1 of the JSONLD..  it's a nice ID
		h := sha1.New()
		h.Write([]byte(jsonld))
		bs := h.Sum(nil)
		bss := fmt.Sprintf("%x", bs) // better way to convert bs hex string to string?

		// objectName := fmt.Sprintf("%s/%s.jsonld", up.Path, bss)
		// objectName := fmt.Sprintf("%s.jsonld", bss)
		// objectName := fmt.Sprintf("%s/%s.jsonld", k, bss)
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
		//bucketName := fmt.Sprintf("gleaner-summoned/%s", k) // old was just k

		err = StoreProv(v1, mc, k, sha, url)
		if err != nil {
			log.Println(err)
		}

		// Upload the  file with FPutObject
		n, err := mc.PutObject(context.Background(), bucketName, objectName, b, int64(b.Len()), minio.PutObjectOptions{ContentType: contentType, UserMetadata: usermeta})
		if err != nil {
			log.Printf("%s", objectName)
			log.Println(err)
		}
		log.Printf("Uploaded Bucket:%s File:%s Size %d \n", bucketName, objectName, n.Size)
	}

	// // Fetch the document root node. We can pass nil here
	// // since this method only takes optional arguments.
	// doc, err := c.DOM.GetDocument(ctx, nil)
	// if err != nil {
	// 	return err
	// }

	// // Get the outer HTML for the page.
	// // #jsonld
	// // document.querySelector("#jsonld")
	// // //*[@id="jsonld"]
	// // /html/head/script[5]
	// result, err := c.DOM.GetOuterHTML(ctx, &dom.GetOuterHTMLArgs{
	// 	NodeID: &doc.Root.NodeID,
	// })
	// if err != nil {
	// 	return err
	// }

	// fmt.Printf("HTML: %s\n", len(result.OuterHTML))

	return nil
}

// runBatchFunc is the function signature for runBatch.
type runBatchFunc func() error

// runBatch runs all functions simultaneously and waits until
// execution has completed or an error is encountered.
func runBatch(fn ...runBatchFunc) error {
	eg := errgroup.Group{}
	for _, f := range fn {
		eg.Go(f)
	}
	return eg.Wait()
}

// Code below here is not used

// setCookies sets all the provided cookies.
func setCookies(ctx context.Context, net cdp.Network, cookies ...Cookie) error {
	var cmds []runBatchFunc
	for _, c := range cookies {
		args := network.NewSetCookieArgs(c.Name, c.Value).SetURL(c.URL)
		cmds = append(cmds, func() error {
			reply, err := net.SetCookie(ctx, args)
			if err != nil {
				return err
			}
			if !reply.Success {
				return errors.New("could not set cookie")
			}
			return nil
		})
	}
	return runBatch(cmds...)
}

// navigate to the URL and wait for DOMContentEventFired. An error is
// returned if timeout happens before DOMContentEventFired.
func navigate(ctx context.Context, pageClient cdp.Page, url string, timeout time.Duration) error {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()

	// Make sure Page events are enabled.
	err := pageClient.Enable(ctx)
	if err != nil {
		return err
	}

	// Open client for DOMContentEventFired to block until DOM has fully loaded.
	domContentEventFired, err := pageClient.DOMContentEventFired(ctx)
	if err != nil {
		return err
	}
	defer domContentEventFired.Close()

	_, err = pageClient.Navigate(ctx, page.NewNavigateArgs(url))
	if err != nil {
		return err
	}

	_, err = domContentEventFired.Recv()
	return err
}

func abortOnErrors(ctx context.Context, c *cdp.Client, abort chan<- error) error {
	exceptionThrown, err := c.Runtime.ExceptionThrown(ctx)
	if err != nil {
		return err
	}

	loadingFailed, err := c.Network.LoadingFailed(ctx)
	if err != nil {
		return err
	}

	go func() {
		defer exceptionThrown.Close() // Cleanup.
		defer loadingFailed.Close()
		for {
			select {
			// Check for exceptions so we can abort as soon
			// as one is encountered.
			case <-exceptionThrown.Ready():
				ev, err := exceptionThrown.Recv()
				if err != nil {
					// This could be any one of: stream closed,
					// connection closed, context deadline or
					// unmarshal failed.
					abort <- err
					return
				}

				// Ruh-roh! Let the caller know something went wrong.
				abort <- ev.ExceptionDetails

			// Check for non-canceled resources that failed
			// to load.
			case <-loadingFailed.Ready():
				ev, err := loadingFailed.Recv()
				if err != nil {
					abort <- err
					return
				}

				// For now, most optional fields are pointers
				// and must be checked for nil.
				canceled := ev.Canceled != nil && *ev.Canceled

				if !canceled {
					abort <- fmt.Errorf("request %s failed: %s", ev.RequestID, ev.ErrorText)
				}
			}
		}
	}()
	return nil
}
