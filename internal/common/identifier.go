package common

/* info on possible packages:
https://cburgmer.github.io/json-path-comparison/
using https://github.com/ohler55/ojg

test your jsonpaths here:
http://jsonpath.herokuapp.com/
There are four implementations... so you can see if one might be a little quirky
*/
import (
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
)

func GetIdentifierByPath(jsonPath string, jsonld string) (interface{}, error) {
	obj, err := oj.ParseString(jsonld)

	x, err := jp.ParseString(jsonPath)
	ys := x.Get(obj)

	if err != nil {
		return "", err
	}
	return ys, err
}
