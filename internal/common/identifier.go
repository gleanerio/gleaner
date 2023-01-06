package common

/* info on possible packages:
https://cburgmer.github.io/json-path-comparison/
using https://github.com/ohler55/ojg

test your jsonpaths here:
http://jsonpath.herokuapp.com/
There are four implementations... so you can see if one might be a little quirky
*/
import (
	"errors"
	"fmt"
	"github.com/gleanerio/gleaner/internal/config"
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Identifier struct {
	uniqueId       string // file sha, identifier sha, or url normalized identifier
	identifierType string // the returned identifierType..
	matchedPath    string
	matchedString  string
}

var jsonPathsDefault = []string{"$['@graph'][?(@['@type']=='schema:Dataset')]['@id']", "$.identifier[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value", "$.identifier.value", "$.identifier", `$['@id']`}

func GenerateIdentifier(v1 *viper.Viper, source config.Sources, jsonld string) (Identifier, error) {

	// Generate calls also do the casecading aka if IdentifierSha is [] it calls Filesha
	switch source.IdentifierType {
	case config.IdentifierString:
		return Identifier{}, errors.New("Not implemented")
	case config.IdentifierSha:
		return GenerateIdentifierSha(v1, source, jsonld)
	default: //config.filesha
		return GenerateFileSha(v1, jsonld)

	}

}

func GetIdentifierByPath(jsonPath string, jsonld string) (interface{}, error) {
	obj, err := oj.ParseString(jsonld)
	if err != nil {
		return "", err
	}
	x, err := jp.ParseString(jsonPath)
	ys := x.Get(obj)

	if err != nil {
		return "", err
	}
	return ys, err
}

// given a set of json paths return the first to the last.
/*
Pass an array of JSONPATH, and get returned the first not empty, result
Cautions: test your paths, consensus returns [] for a $.identifer.value, even through

{ identifier:"string"}
has no value:

"idenfitier":
	{
	"@type": "PropertyValue",
	"@id": "https://doi.org/10.1575/1912/bco-dmo.2343.1",
	"propertyID": "https://registry.identifiers.org/registry/doi",
	"value": "doi:10.1575/1912/bco-dmo.2343.1",
	"url": "https://doi.org/10.1575/1912/bco-dmo.2343.1"
	}
https://cburgmer.github.io/json-path-comparison/results/dot_notation_on_object_without_key.html
https://cburgmer.github.io/json-path-comparison/results/dot_notation_on_null_value.html
*/
func GetIdentiferByPaths(jsonpaths []string, jsonld string) (interface{}, string, error) {
	for _, jsonPath := range jsonpaths {
		obj, err := GetIdentifierByPath(jsonPath, jsonld)
		if err == nil {
			// returned a string, but
			// sometimes an empty string is returned
			if fmt.Sprint(obj) == "[]" {
				continue
			}
			return obj, jsonPath, err

		} else {
			// error,
			continue
		}
	}
	return "", "", errors.New("No Match")
}

func GenerateIdentifierSha(v1 *viper.Viper, source config.Sources, jsonld string) (Identifier, error) {
	// need a copy of the arrays, or it will get munged in a multithreaded env
	var jsonpath = make([]string, len(jsonPathsDefault))
	copy(jsonpath, jsonPathsDefault)
	if len(source.IdentifierPath) > 0 && source.IdentifierPath != nil {
		// this does not move an item to the front of the array, if the item already exists in the array,
		// overriding the default overrides all paths
		//jsonpath = append(source.IdentifierPath, jsonPathsDefault...)
		//jsonpath = source.IdentifierPath
		for _, p := range source.IdentifierPath {
			jsonpath = config.MoveToFront(p, jsonPathsDefault)
		}

	}
	uniqueid, foundPath, err := GetIdentiferByPaths(jsonpath, jsonld)
	if err == nil && uniqueid != "[]" {
		id := Identifier{uniqueId: GetSHA(fmt.Sprint(uniqueid)),
			identifierType: config.IdentifierSha,
			matchedPath:    foundPath,
			matchedString:  fmt.Sprint(uniqueid)}
		return id, err
	} else {
		log.Error(config.IdentifierSha, "Action: Getting normalized sha  Error:", err)
		// generate a filesha
		return GenerateFileSha(v1, jsonld)
	}
}
func GenerateFileSha(v1 *viper.Viper, jsonld string) (Identifier, error) {

	//uuid := common.GetSHA(jsonld)
	uuid, err := GetNormSHA(jsonld, v1) // Moved to the normalized sha value
	if err != nil {
		log.Error("ERROR: uuid generator:", "Action: Getting normalized sha  Error:", err)
		return Identifier{}, err
	}
	id := Identifier{uniqueId: uuid,
		identifierType: config.Filesha,
	}
	log.Info("filesha: ", uuid)
	fmt.Println("\nfilesha:", id)
	return id, err
}
