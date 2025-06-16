package graph

import (
	"github.com/piprate/json-gold/ld"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	"os"
)

// ContextMapping holds the JSON-LD mappings for cached context
type ContextMapping struct {
	Prefix string
	File   string
}

// JLDProc builds the JSON-LD processor and sets the options object
// for use in framing, processing and all JSON-LD actions
func JLDProc(v1 *viper.Viper) (*ld.JsonLdProcessor, *ld.JsonLdOptions) {
	proc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")

	mcfg := v1.GetStringMapString("context")

	if mcfg["cache"] == "true" {
		client := &http.Client{}
		nl := ld.NewDefaultDocumentLoader(client)

		var s []ContextMapping
		err := v1.UnmarshalKey("contextmaps", &s)
		if err != nil {
			log.Error(err)
		}

		m := make(map[string]string)

		for i := range s {
			if fileExists(s[i].File) {
				m[s[i].Prefix] = s[i].File

			} else {
				log.Error("ERROR: context file location ", s[i].File, " is wrong, this is a critical error")
			}
		}

		// Read mapping from config file
		cdl := ld.NewCachingDocumentLoader(nl)
		err = cdl.PreloadWithMapping(m)
		if err != nil {
			return nil, nil
		}
		options.DocumentLoader = cdl
	}

	options.ProcessingMode = ld.JsonLd_1_1 // add mode explicitly if you need JSON-LD 1.1 features
	options.Format = "application/nquads"  // Set to a default format. (make an option?)

	return proc, options
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
