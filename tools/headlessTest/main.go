// Command text is a chromedp example demonstrating how to extract text from a
// specific element.
package main

import (
	"context"
	"log"
	"strings"

	"github.com/chromedp/chromedp"
)

func main() {
	// create context
	log.SetFlags(log.LstdFlags | log.Llongfile)
	ctx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()

	// run task list

	// https://github.com/chromedp/chromedp/issues/507
	//document.querySelector("#jsonld")
	var res string
	err := chromedp.Run(ctx,
		chromedp.Navigate(`http://igsn.org/ICDP5054EC2U001`),
		chromedp.Text(`#jsonld`, &res, chromedp.NodeVisible, chromedp.ByID),
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(strings.TrimSpace(res))
}
