package acquire

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"reflect"
	"strings"

	"github.com/gleanerio/gleaner/internal/common"
	minio "github.com/minio/minio-go/v7"
	"github.com/spf13/viper"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// / A utility to keep a list of JSON-LD files that we have found
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

// / Validate JSON-LD that we get
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

// *********
// context fixes
// *********
// let's try to do them all, in one, since that will make the code a bit cleaner and easier to test
const httpContext = "http://schema.org/"
const httpsContext = "https://schema.org/"

type ContextOption int64

const (
	Strict ContextOption = iota
	Https
	Http
	//	Array
	//	Object
	StandardizedHttps
	StandardizedHttp
)

func (s ContextOption) String() string {
	switch s {
	case Strict:
		return "strict"
	case Https:
		return "https"
	case Http:
		return "http"
		//	case Array:
		//		return "array"
		//	case Object:
		//		return "object"
	case StandardizedHttps:
		return "standardizedHttps"
	case StandardizedHttp:
		return "standardizedHttp"
	}
	return "unknown"
}

func fixContext(jsonld string, option ContextOption) (string, error) {
	var err error

	if option == Strict {
		return jsonld, nil
	}
	jsonContext := gjson.Get(jsonld, "@context")

	var ctxSchemaOrg = httpsContext
	if option == Http {
		ctxSchemaOrg = httpContext
	}

	//switch jsonContext.Value().(type) {
	switch reflect.ValueOf(jsonContext).Kind() {
	//case string:
	case reflect.String:
		jsonld, err = fixContextString(jsonld, Https)
	case reflect.Slice:
		jsonld, err = fixContextArray(jsonld, Https)
	}
	jsonld, err = fixContextUrl(jsonld, ctxSchemaOrg)
	return jsonld, err
}

// Our first json fixup in existence.
// If the top-level JSON-LD context is a string instead of an object,
// this function corrects it.
func fixContextString(jsonld string, option ContextOption) (string, error) {
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
// this needs to check all items to see if they match schema.org, then fix.
func fixContextUrl(jsonld string, ctx string) (string, error) {
	var err error
	//context := gjson.Get(jsonld, "@context.@vocab").String()

	contexts := gjson.Get(jsonld, "@context").Map()
	if _, ok := contexts["@vocab"]; !ok {
		jsonld, err = sjson.Set(jsonld, "@context.@vocab", httpsContext)
	}
	// for range
	for ns, c := range contexts {
		var context = c.String()
		//var changed = false
		if strings.Contains(context, "schema.org") {
			//if !strings.HasSuffix(context, "/") {
			//	context += "/"
			//	changed = true
			//}
			//contextUrl, _ := url.Parse(context)
			//if contextUrl.Scheme != "https" {
			//	contextUrl.Scheme = "https"
			//	context = contextUrl.String()
			//	changed = true
			//}
			if strings.Contains(context, "www.") { // fix www.schema.org
				var i = strings.Index(context, "schema.org")
				context = context[i:]
				context = ctx + context
			}
			if len(context) < 20 { // https://schema.org/
				context = ctx
			}

			//changed=true
		}
		//if changed {
		// contexts[ns] = context
		var path = "@context." + ns
		jsonld, err = sjson.Set(jsonld, path, context)
		if err != nil {
			log.Error("Error standardizing schema.org" + err.Error())
		}
		//}

	}
	//jsonld, err = sjson.Set(jsonld, "@context", map[string]interface{}{"@vocab": context})
	return jsonld, err
}

// Our first json fixup in existence.
// If the top-level JSON-LD context is a string instead of an object,
// this function corrects it.
func fixContextArray(jsonld string, option ContextOption) (string, error) {
	var err error
	jsonContext := gjson.Get(jsonld, "@context")
	//contexts := gjson.Get(jsonld, "@context").Map()
	switch jsonContext.Value().(type) {
	//switch reflect.ValueOf(contexts).Kind() {
	//case string:
	//case reflect.Slice:
	case nil:
		jsonld, err = sjson.Set(jsonld, "@context", map[string]interface{}{"@vocab": httpsContext})
	}
	return jsonld, err
}

func Upload(v1 *viper.Viper, mc *minio.Client, bucketName string, site string, urlloc string, jsonld string) (string, error) {
	mcfg := v1.GetStringMapString("context")
	var err error

	// In the config file, context { strict: true } bypasses these fixups.
	// Strict defaults to false.
	if strict, ok := mcfg["strict"]; !(ok && strict == "true") {
		log.Info("context.strict is not set to true; doing json-ld fixups.")
		//contextType := reflect.ValueOf(jsonld).Kind().String()
		//if strings.HasPrefix(contextType, "string") || strings.HasPrefix(contextType, "map") {
		//
		//}
		jsonld, err = fixContextString(jsonld, Https)
		if err != nil {
			log.Error("ERROR: URL:", urlloc, "Action: Fixing JSON-LD context to be an object Error:", err)
		}
		//jsonld, err = fixContextUrl(jsonld, Https)
		jsonld, err = fixContextUrl(jsonld, httpsContext) // CONST for now
		if err != nil {
			log.Error("ERROR: URL:", urlloc, "Action: Fixing JSON-LD context url scheme and trailing slash Error:", err)
		}
	}
	sha, err := common.GetNormSHA(jsonld, v1) // Moved to the normalized sha value
	if err != nil {
		log.Error("ERROR: URL:", urlloc, "Action: Getting normalized sha  Error:", err)
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
		log.Error(err)
	}

	// Upload the file with FPutObject
	_, err = mc.PutObject(context.Background(), bucketName, objectName, b, int64(b.Len()), minio.PutObjectOptions{ContentType: contentType, UserMetadata: usermeta})
	if err != nil {
		log.Fatal(objectName, err) // Fatal?   seriously?    I guess this is the object write, so the run is likely a bust at this point, but this seems a bit much still.
	}
	log.Debug("Uploaded Bucket:", bucketName, " File:", objectName, "Size", int64(b.Len()))
	return sha, err
}
