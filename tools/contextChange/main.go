package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	//"github.com/gleanerio/gleaner/internal/summoner/acquire"
)

// Simple test of the nanoprov function
func main() {

	// Read file
	fmt.Println("Read in a JSON-LD data graph and test changing its context")
	dat, err := os.ReadFile("../../secret/data/sprep.json")
	if err != nil {
		fmt.Println("Error reading file")
	}

	jld, err := fixContextUrl(string(dat))
	if err != nil {
		fmt.Println("Error reading file")
	}

	fmt.Println(jld)

	// pass to context change

}

// Our first json fixup in existence.
// If the top-level JSON-LD context is a string instead of an object,
// this function corrects it.
func fixContextString(jsonld string) (string, error) {
	var err error
	jsonContext := gjson.Get(jsonld, "@context")

	switch jsonContext.Value().(type) {
	case string:
		jsonld, err = sjson.Set(jsonld, "@context", map[string]interface{}{"@vocab": jsonContext.String()})
	}
	return jsonld, err
}

// If the top-level JSON-LD context does not end with a trailing slash or use https,
// this function corrects it.
func fixContextUrl(jsonld string) (string, error) {
	var err error
	context := gjson.Get(jsonld, "@context.@vocab").String()
	if !strings.HasSuffix(context, "/") {
		context += "/"
	}
	contextUrl, err := url.Parse(context)
	if contextUrl.Scheme != "https" {
		contextUrl.Scheme = "https"
		context = contextUrl.String()
	}

	jsonld, err = sjson.Set(jsonld, "@context", map[string]interface{}{"@vocab": context})
	return jsonld, err
}
