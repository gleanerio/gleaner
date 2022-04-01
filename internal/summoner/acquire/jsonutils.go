package acquire

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/url"
	"strings"

	"github.com/gleanerio/gleaner/internal/common"
	minio "github.com/minio/minio-go/v7"
	"github.com/spf13/viper"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

/// A utility to keep a list of JSON-LD files that we have found
// in or on a page
func addToJsonListIfValid(v1 *viper.Viper, jsonlds []string, new_json string) ([]string, error) {
	valid, err := isValid(v1, new_json)
	if err != nil {
		return jsonlds, fmt.Errorf("error checking for valid json: %s", err)
	}
	if !valid {
		return jsonlds, fmt.Errorf("invalid json; continuing")
	}
	return append(jsonlds, new_json), nil
}

/// Validate JSON-LD that we get
func isValid(v1 *viper.Viper, jsonld string) (bool, error) {
	proc, options := common.JLDProc(v1)

	var myInterface map[string]interface{}

	err := json.Unmarshal([]byte(jsonld), &myInterface)
	if err != nil {
		return false, fmt.Errorf("Error in unmarshaling json: %s", err)
	}

	_, err = proc.ToRDF(myInterface, options) // returns triples but toss them, just validating
	if err != nil {                           // it's wasted cycles.. but if just doing a summon, needs to be done here
		return false, fmt.Errorf("Error in JSON-LD to RDF call: %s", err)
	}

	return true, nil
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
	if ! strings.HasSuffix(context, "/") {
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

func Upload(v1 *viper.Viper, mc *minio.Client, logger *log.Logger, bucketName string, site string,  urlloc string, jsonld string) (string, error) {
	mcfg := v1.GetStringMapString("context")

	// In the config file, context { strict: true } bypasses these fixups.
	// Strict defaults to false.
	if strict, ok := mcfg["strict"]; !(ok && strict == "true") {
		log.Println("context.strict is not set to true; doing json-ld fixups.")
		jsonld, err := fixContextString(jsonld)
		if err != nil {
			log.Printf("ERROR: URL: %s Action: Fixing JSON-LD context to be an object Error: %s\n", urlloc, err)
		}
		jsonld, err = fixContextUrl(jsonld)
		if err != nil {
			log.Printf("ERROR: URL: %s Action: Fixing JSON-LD context url scheme and trailing slash Error: %s\n", urlloc, err)
		}
	}
	sha, err := common.GetNormSHA(jsonld, v1) // Moved to the normalized sha value
	if err != nil {
		log.Printf("ERROR: URL: %s Action: Getting normalized sha  Error: %s\n", urlloc, err)
	}
	objectName := fmt.Sprintf("summoned/%s/%s.jsonld", site, sha)
	contentType := "application/ld+json"
	b := bytes.NewBufferString(jsonld)
	// size := int64(b.Len()) // gets set to 0 after upload for some reason

	usermeta := make(map[string]string) // what do I want to know?
	usermeta["url"] = urlloc
	usermeta["sha1"] = sha

	// write the prov entry for this object
	err = StoreProvNG(v1, mc, site, sha, urlloc, "milled")
	if err != nil {
		log.Println(err)
	}

	// Upload the file with FPutObject
	_, err = mc.PutObject(context.Background(), bucketName, objectName, b, int64(b.Len()), minio.PutObjectOptions{ContentType: contentType, UserMetadata: usermeta})
	if err != nil {
		log.Printf("%s", objectName)
		log.Fatalln(err) // Fatal?   seriously?    I guess this is the object write, so the run is likely a bust at this point, but this seems a bit much still.
	}

	return sha, err

	// log.Printf("Uploaded Bucket:%s File:%s Size %d\n", bucketName, objectName, size)
}
