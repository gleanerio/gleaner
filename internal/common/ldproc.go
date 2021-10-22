package common

import (
	"log"
	"net/http"
	"os"

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
// TODO   we create this all the time..  stupidly..  Generate these pointers
// and pass them around, don't keep making it over and over
// Ref:  https://schema.org/docs/howwework.html and https://schema.org/docs/jsonldcontext.json
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
			if fileExists(s[i].File) {
				m[s[i].Prefix] = s[i].File

			} else {
				log.Printf("ERROR: context file location %s is wrong, this is a critical error", s[i].File)
			}
		}

		// Read mapping from config file
		cdl := ld.NewCachingDocumentLoader(nl)
		cdl.PreloadWithMapping(m)
		options.DocumentLoader = cdl
		// todo: check domain config and see whether it should be processed with 1.0 options.ProcessingMode = "json-ld-1.0"
	}

	// Set a default format..  let this be set later...
	options.Format = "application/nquads"

	return proc, options
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
