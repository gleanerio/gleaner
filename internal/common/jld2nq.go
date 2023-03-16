package common

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"

	"github.com/piprate/json-gold/ld"
)

// JLD2nq converts JSON-LD documents to NQuads
func JLD2nq(jsonld string, proc *ld.JsonLdProcessor, options *ld.JsonLdOptions) (string, error) {
	var myInterface interface{}
	err := json.Unmarshal([]byte(jsonld), &myInterface)
	if err != nil {
		log.Error(err)
		return "", err
	}

	nq, err := proc.ToRDF(myInterface, options)
	if err != nil {
		log.Error(err)

		return "", err
	}

	return nq.(string), err
}
