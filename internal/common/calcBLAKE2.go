package common

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/viper"
	"golang.org/x/crypto/blake2b"
)

func GetBLAKE2(jsonld string, v1 *viper.Viper) (string, error) {
	// proc, options := JLDProc(v1)

	// // proc := ld.NewJsonLdProcessor()
	// // options := ld.NewJsonLdOptions("")
	// // add the processing mode explicitly if you need JSON-LD 1.1 features
	// options.ProcessingMode = ld.JsonLd_1_1
	// options.Format = "application/n-quads"
	// options.Algorithm = "URDNA2015"

	// // JSON-LD   this needs to be an interface, otherwise it thinks it is a URL to get
	// var myInterface interface{}
	// err := json.Unmarshal([]byte(jsonld), &myInterface)
	// if err != nil {
	// 	return "", err
	// }

	// normalizedTriples, err := proc.Normalize(myInterface, options)
	// if err != nil {
	// 	return "", err
	// }

	// h := blake2b.Sum256([]byte(normalizedTriples.(string)))

	h := blake2b.Sum256([]byte(jsonld))

	return fmt.Sprintf("%x", hex.EncodeToString(h[:])), nil
}
