package acquire

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gleanerio/gleaner/internal/config"
	log "github.com/sirupsen/logrus"
	"net/url"
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
// don' think this is currently called anywhere
const httpContext = "http://schema.org/"
const httpsContext = "https://schema.org/"

func fixContext(jsonld string, option config.ContextOption) (string, error) {
	var err error

	if option == config.Strict {
		return jsonld, nil
	}
	jsonContext := gjson.Get(jsonld, "@context")

	var ctxSchemaOrg = httpsContext
	if option == config.Http {
		ctxSchemaOrg = httpContext
	}

	switch reflect.ValueOf(jsonContext).Kind() {
	case reflect.String:
		jsonld, err = fixContextString(jsonld, config.Https)
	case reflect.Slice:
		jsonld, err = fixContextArray(jsonld, config.Https)
	}
	jsonld, err = fixContextUrl(jsonld, ctxSchemaOrg)
	return jsonld, err
}

// Our first json fixup in existence.
// If the top-level JSON-LD context is a string instead of an object,
// this function corrects it.
func fixContextString(jsonld string, option config.ContextOption) (string, error) {
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
	contexts := gjson.Get(jsonld, "@context").Map()
	if _, ok := contexts["@vocab"]; !ok {
		jsonld, err = sjson.Set(jsonld, "@context.@vocab", httpsContext)
	}
	// for range
	for ns, c := range contexts {
		var context = c.String()
		if strings.Contains(context, "schema.org") {
			if strings.Contains(context, "www.") { // fix www.schema.org
				var i = strings.Index(context, "schema.org")
				context = context[i:]
				context = ctx + context
			}
			if len(context) < 20 { // https://schema.org/
				context = ctx
			}
		}
		var path = "@context." + ns
		jsonld, err = sjson.Set(jsonld, path, context)
		if err != nil {
			log.Error("Error standardizing schema.org" + err.Error())
		}

	}
	return jsonld, err
}

// Our first json fixup in existence.
// If the top-level JSON-LD context is a string instead of an object,
// this function corrects it.
func fixContextArray(jsonld string, option config.ContextOption) (string, error) {
	var err error
	contexts := gjson.Get(jsonld, "@context")
	switch contexts.Value().(type) {
	case []interface{}: // array
		jsonld, err = standardizeContext(jsonld, config.StandardizedHttps)
	case map[string]interface{}: // array
		jsonld = jsonld
	}
	return jsonld, err
}

// if the top-level JSON-LD @id is not an IRI, and there is no base in the context,
// remove that id
// see https://github.com/piprate/json-gold/discussions/68#discussioncomment-4782788
// for details
func fixId(jsonld string) (string, error) {
	var err error
	originalBase := gjson.Get(jsonld, "@context.@base").String()
	if originalBase != "" { // if we have a context base, there is no need to do any of this
		return jsonld, err
	}
	topLevelType := gjson.Get(jsonld, "@type").String()
	var selector string
	var formatter func(index int) string
	if topLevelType == "Dataset" {
		selector = "@id"
		formatter = func(index int) string { return "@id"}
	} else if topLevelType == "ItemList" {
		selector = "itemListElement.#.item.@id"
		formatter = func(index int) string { return fmt.Sprintf("itemListElement.%v.item.@id", index) }
	} else { // we don't know how to fix any of these other things
		log.Trace("Found a top-level type of ", topLevelType, " in this jsonld document")
		return jsonld, err
	}
	jsonIdentifiers := gjson.Get(jsonld, selector)
	index := 0
	jsonIdentifiers.ForEach(func(key, jsonResult gjson.Result) bool {
		jsonIdentifier := jsonResult.String()
		idUrl, idErr := url.Parse(jsonIdentifier)
		if idUrl.Scheme == "" { // we have a relative url and no base in the context
			log.Trace("Transforming id: ", jsonIdentifier, " to file:// url because it is relative")
			jsonld, idErr = sjson.Set(jsonld, formatter(index), "file://" + jsonIdentifier)
		} else {
			log.Trace("JSON-LD context base or IRI id found: ", originalBase, "ID: ", idUrl)
		}
		if idErr != nil {
			err = idErr
			return false
		}
		index++
		return true
	})
	return jsonld, err
}

// this just creates a standardized context
// jsonMap := make(map[string]interface{})
var StandardHttpsContext = map[string]interface{}{
	"@vocab": "https://schema.org/",
	"adms":   "https://www.w3.org/ns/adms#",
	"dcat":   "https://www.w3.org/ns/dcat#",
	"dct":    "https://purl.org/dc/terms/",
	"foaf":   "https://xmlns.com/foaf/0.1/",
	"gsp":    "https://www.opengis.net/ont/geosparql#",
	"locn":   "https://www.w3.org/ns/locn#",
	"owl":    "https://www.w3.org/2002/07/owl#",
	"rdf":    "https://www.w3.org/1999/02/22-rdf-syntax-ns#",
	"rdfs":   "https://www.w3.org/2000/01/rdf-schema#",
	"schema": "https://schema.org/",
	"skos":   "https://www.w3.org/2004/02/skos/core#",
	"spdx":   "https://spdx.org/rdf/terms#",
	"time":   "https://www.w3.org/2006/time",
	"vcard":  "https://www.w3.org/2006/vcard/ns#",
	"xsd":    "https://www.w3.org/2001/XMLSchema#",
}

var StandardHttpContext = map[string]interface{}{
	"@vocab": "http://schema.org/",
	"adms":   "http://www.w3.org/ns/adms#",
	"dcat":   "http://www.w3.org/ns/dcat#",
	"dct":    "http://purl.org/dc/terms/",
	"foaf":   "http://xmlns.com/foaf/0.1/",
	"gsp":    "http://www.opengis.net/ont/geosparql#",
	"locn":   "http://www.w3.org/ns/locn#",
	"owl":    "http://www.w3.org/2002/07/owl#",
	"rdf":    "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
	"rdfs":   "http://www.w3.org/2000/01/rdf-schema#",
	"schema": "http://schema.org/",
	"skos":   "http://www.w3.org/2004/02/skos/core#",
	"spdx":   "http://spdx.org/rdf/terms#",
	"time":   "http://www.w3.org/2006/time",
	"vcard":  "http://www.w3.org/2006/vcard/ns#",
	"xsd":    "http://www.w3.org/2001/XMLSchema#",
}

func standardizeContext(jsonld string, option config.ContextOption) (string, error) {

	var err error

	switch option {
	case config.StandardizedHttps:
		jsonld, err = sjson.Set(jsonld, "@context", StandardHttpsContext)
	case config.StandardizedHttp:
		jsonld, err = sjson.Set(jsonld, "@context", StandardHttpContext)
	}
	return jsonld, err
}

// there is a cleaner way to handle this...
func getOptions(ctxOption config.ContextOption) (config.ContextOption, string) {
	var fixTpye = config.Https
	var ctxString = httpsContext
	if ctxOption != config.Strict {
		if ctxOption == config.Https || ctxOption == config.StandardizedHttps {
			fixTpye = config.Https
			ctxString = httpsContext
		} else {
			fixTpye = config.Http
			ctxString = httpContext
		}
		return fixTpye, ctxString
	} else {
		return config.Strict, ctxString
	}

}

// ##### end contxt fixes
func ProcessJson(v1 *viper.Viper,
	source *config.Sources, urlloc string, jsonld string) (string, common.Identifier, error) {
	mcfg := v1.GetStringMapString("context")
	var err error
	//sources, err := config.GetSources(v1)
	//source, err := config.GetSourceByName(sources, site)
	srcFixOption, srcHttpOption := getOptions(source.FixContextOption)

	// In the config file, context { strict: true } bypasses these fixups.
	// Strict defaults to false.
	// this is a command level
	if strict, ok := mcfg["strict"]; !(ok && strict == "true") || (srcFixOption != config.Strict) {
		// source level

		log.Info("context.strict is not set to true; doing json-ld fixups.")
		jsonld, err = fixContextString(jsonld, srcFixOption)
		if err != nil {
			log.Error(
				"ERROR: URL: ", urlloc, " Action: Fixing JSON-LD context from string to be an object Error: ", err)
		}
		jsonld, err = fixContextArray(jsonld, srcFixOption)
		if err != nil {
			log.Error("ERROR: URL: ", urlloc, " Action: Fixing JSON-LD context from array to be an object Error: ", err)
		}
		jsonld, err = fixContextUrl(jsonld, srcHttpOption) // CONST for now
		if err != nil {
			log.Error("ERROR: URL: ", urlloc, " Action: Fixing JSON-LD context url scheme and trailing slash Error: ", err)
		}
		jsonld, err = fixId(jsonld)
		if err != nil {
			log.Error("ERROR: URL: ", urlloc, " Action: Removing relative JSON-LD @id Error: ", err)
		}

	}
	//sha, err := common.GetNormSHA(jsonld, v1) // Moved to the normalized sha value
	identifier, err := common.GenerateIdentifier(v1, *source, jsonld)
	if err != nil {
		log.Error("ERROR: URL:", urlloc, "Action: Getting normalized sha  Error:", err)
	}
	//sha := identifier.UniqueId
	//objectName := fmt.Sprintf("summoned/%s/%s.jsonld", site, sha)
	//contentType := "application/ld+json"
	//b := bytes.NewBufferString(jsonld)
	//// size := int64(b.Len()) // gets set to 0 after upload for some reason
	//
	//usermeta := make(map[string]string) // what do I want to know?
	//usermeta["url"] = urlloc
	//usermeta["sha1"] = sha
	//usermeta["uniqueid"] = sha
	//usermeta["jsonsha"] = identifier.JsonSha
	//usermeta["identifiertype"] = identifier.IdentifierType
	//if identifier.MatchedPath != "" {
	//	usermeta["matchedpath"] = identifier.MatchedPath
	//	usermeta["matchedstring"] = identifier.MatchedString
	//}
	//if config.IdentifierString == source.IdentifierType {
	//	usermeta["sha1"] = identifier.JsonSha
	//}
	//if source.IdentifierType == config.SourceUrl {
	//	log.Info("not suppported, yet. needs url sanitizing")
	//}
	// write the prov entry for this object
	//err = StoreProvNG(v1, mc, site, sha, urlloc, "milled")
	//if err != nil {
	//	log.Error(err)
	//}
	//
	//// ProcessJson the file with FPutObject
	//_, err = mc.PutObject(context.Background(), bucketName, objectName, b, int64(b.Len()), minio.PutObjectOptions{ContentType: contentType, UserMetadata: usermeta})
	//if err != nil {
	//	log.Fatal(objectName, err) // Fatal?   seriously?    I guess this is the object write, so the run is likely a bust at this point, but this seems a bit much still.
	//}
	//log.Debug("Uploaded Bucket:", bucketName, " File:", objectName, "Size", int64(b.Len()))
	return jsonld, identifier, err
}

func Upload(v1 *viper.Viper, mc *minio.Client, bucketName string, site string, urlloc string, jsonld string) (string, error) {
	var err error
	//mcfg := v1.GetStringMapString("context")
	sources, err := config.GetSources(v1)
	source, err := config.GetSourceByName(sources, site)
	//srcFixOption, srcHttpOption := getOptions(source.FixContextOption)

	//// In the config file, context { strict: true } bypasses these fixups.
	//// Strict defaults to false.
	//// this is a command level
	//if strict, ok := mcfg["strict"]; !(ok && strict == "true") || (srcFixOption != config.Strict) {
	//	// source level
	//
	//	log.Info("context.strict is not set to true; doing json-ld fixups.")
	//	//contextType := reflect.ValueOf(jsonld).Kind().String()
	//	//if strings.HasPrefix(contextType, "string") || strings.HasPrefix(contextType, "map") {
	//	//
	//	//}
	//	jsonld, err = fixContextString(jsonld, srcFixOption)
	//	if err != nil {
	//		log.Error("ERROR: URL:", urlloc, "Action: Fixing JSON-LD context from string to be an object Error:", err)
	//	}
	//	jsonld, err = fixContextArray(jsonld, srcFixOption)
	//	if err != nil {
	//		log.Error("ERROR: URL:", urlloc, "Action: Fixing JSON-LD context from array to be an object Error:", err)
	//	}
	//	//jsonld, err = fixContextUrl(jsonld, Https)
	//	jsonld, err = fixContextUrl(jsonld, srcHttpOption) // CONST for now
	//	if err != nil {
	//		log.Error("ERROR: URL:", urlloc, "Action: Fixing JSON-LD context url scheme and trailing slash Error:", err)
	//	}
	//
	//}
	////sha, err := common.GetNormSHA(jsonld, v1) // Moved to the normalized sha value
	//identifier, err := common.GenerateIdentifier(v1, *source, jsonld)
	//if err != nil {
	//	log.Error("ERROR: URL:", urlloc, "Action: Getting normalized sha  Error:", err)
	//}
	jsonld, identifier, err := ProcessJson(v1, source, urlloc, jsonld)

	sha := identifier.UniqueId
	objectName := fmt.Sprintf("summoned/%s/%s.jsonld", site, sha)
	contentType := JSONContentType
	b := bytes.NewBufferString(jsonld)
	// size := int64(b.Len()) // gets set to 0 after upload for some reason

	usermeta := make(map[string]string) // what do I want to know?
	usermeta["url"] = urlloc
	usermeta["sha1"] = sha
	usermeta["uniqueid"] = sha
	usermeta["jsonsha"] = identifier.JsonSha
	usermeta["identifiertype"] = identifier.IdentifierType
	if identifier.MatchedPath != "" {
		usermeta["matchedpath"] = identifier.MatchedPath
		usermeta["matchedstring"] = identifier.MatchedString
	}
	if config.IdentifierString == source.IdentifierType {
		usermeta["sha1"] = identifier.JsonSha
	}
	if source.IdentifierType == config.SourceUrl {
		log.Info("not suppported, yet. needs url sanitizing")
	}
	// write the prov entry for this object
	err = StoreProvNG(v1, mc, site, sha, urlloc, "milled")
	if err != nil {
		log.Error(err)
	}

	// ProcessJson the file with FPutObject
	_, err = mc.PutObject(context.Background(), bucketName, objectName, b, int64(b.Len()), minio.PutObjectOptions{ContentType: contentType, UserMetadata: usermeta})
	if err != nil {
		log.Fatal(objectName, err) // Fatal?   seriously?    I guess this is the object write, so the run is likely a bust at this point, but this seems a bit much still.
	}
	log.Debug("Uploaded Bucket:", bucketName, " File:", objectName, "Size", int64(b.Len()))
	return sha, err
}
