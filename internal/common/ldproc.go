package common

import (
	"log"
	"net/http"

	"github.com/piprate/json-gold/ld"
	"github.com/spf13/viper"
)

// ContextMapping holds the JSON-LD mappings for cached context
type ContextMapping struct {
	Prefix string
	File   string
}

// JLDProc build the JSON-LD processer and sets the options object
// to use in framing, processing and all JSON-LD actions
func JLDProc(v1 *viper.Viper) (*ld.JsonLdProcessor, *ld.JsonLdOptions) { // TODO make a booklean
	proc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")

	mcfg := v1.GetStringMapString("context")

	if mcfg["cache"] == "true" {
		client := &http.Client{}
		nl := ld.NewDefaultDocumentLoader(client)

		var s []ContextMapping
		err := v1.UnmarshalKey("contextmaps", &s)
		if err != nil {
			log.Println(err)
		}

		m := make(map[string]string)

		for i := range s {
			m[s[i].Prefix] = s[i].File
		}

		// Read mapping from config file
		cdl := ld.NewCachingDocumentLoader(nl)
		cdl.PreloadWithMapping(m)
		options.DocumentLoader = cdl
	}

	// Set a default format..  let this be set later...
	options.Format = "application/nquads"

	return proc, options
}
