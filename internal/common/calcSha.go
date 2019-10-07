package common

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"

	"github.com/piprate/json-gold/ld"
)

// GetSHA returns the sha with any string s
func GetSHA(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	hs := h.Sum(nil)
	return fmt.Sprintf("%x", hs)
}

// GetNormSHA returns the sha hash for a normalized JSON-LD data graph
func GetNormSHA(jsonld string) (string, error) {
	proc, options := JLDProc()

	// proc := ld.NewJsonLdProcessor()
	// options := ld.NewJsonLdOptions("")
	// add the processing mode explicitly if you need JSON-LD 1.1 features
	options.ProcessingMode = ld.JsonLd_1_1
	options.Format = "application/n-quads"
	options.Algorithm = "URDNA2015"

	// JSON-LD   this needs to be an interface, otherwise it thinks it is a URL to get
	var myInterface interface{}
	err := json.Unmarshal([]byte(jsonld), &myInterface)
	if err != nil {
		return "", err
	}

	normalizedTriples, err := proc.Normalize(myInterface, options)
	if err != nil {
		return "", err
	}

	h := sha1.New()
	h.Write([]byte(fmt.Sprint(normalizedTriples.(string))))
	hs := h.Sum(nil)
	return fmt.Sprintf("%x", hs), nil
}
