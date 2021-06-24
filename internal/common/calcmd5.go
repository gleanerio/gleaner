package common

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/piprate/json-gold/ld"
	"github.com/spf13/viper"
)

func GetNormMD5(jsonld string, v1 *viper.Viper) (string, error) {
	proc, options := JLDProc(v1)

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

	r := strings.NewReader(normalizedTriples.(string))

	h := md5.New()
	if _, err := io.Copy(h, r); err != nil {
		log.Fatal(err)
	}

	// h := sha1.New()
	// h.Write([]byte(fmt.Sprint(normalizedTriples.(string))))
	hs := h.Sum(nil)
	return fmt.Sprintf("%x", hs), nil
}
