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
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
)

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
func GetIdentiferByPaths(jsonpaths []string, jsonld string) (interface{}, error) {
	for _, jsonPath := range jsonpaths {
		obj, err := GetIdentifierByPath(jsonPath, jsonld)
		if err != nil {
			continue
		} else {
			return obj, err
		}
	}
	return "", errors.New("No Match")
}
