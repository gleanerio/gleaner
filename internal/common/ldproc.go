package common

import (
	"github.com/piprate/json-gold/ld"
)

// JLDProc build the JSON-LD processer and sets the options object
// to use in framing, processing and all JSON-LD actions
func JLDProc() (*ld.JsonLdProcessor, *ld.JsonLdOptions) {
	proc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")

	// Make a caching client
	// client := &http.Client{}
	// nl := ld.NewDefaultDocumentLoader(client)

	// cdl := ld.NewCachingDocumentLoader(nl)
	// cdl.PreloadWithMapping(map[string]string{"https://schema.org/": "/home/fils/Project418/gleaner/docs/jsonldcontext.json",
	// 	"http://schema.org/": "/home/fils/Project418/gleaner/docs/jsonldcontext.json",
	// 	"https://schema.org": "/home/fils/Project418/gleaner/docs/jsonldcontext.json",
	// 	"http://schema.org":  "/home/fils/Project418/gleaner/docs/jsonldcontext.json"})
	// options.DocumentLoader = cdl

	// Set a default format..  let this be set later...
	options.Format = "application/nquads"

	return proc, options
}
