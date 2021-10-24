package framing

import (
	"encoding/json"
	"github.com/earthcubearchitecture-project418/gleaner/internal/common"
	"github.com/spf13/viper"
	"log"
)

type OrgFrameRes struct {
	Description string
	ID          string
	Type        string
	URL         string
	// URL         string `json:"schema:url"`
}

func OrgFrame(jsonld string, g1 *viper.Viper) []OrgFrameRes {
	proc, options := common.JLDProc(g1)

	// proc := ld.NewJsonLdProcessor()
	// options := ld.NewJsonLdOptions("")

	frame := map[string]interface{}{
		"@context": "http://schema.org/",
		"@type":    "Dataset",
	}

	var myInterface interface{}
	err := json.Unmarshal([]byte(jsonld), &myInterface)
	if err != nil {
		log.Println("Error when transforming JSON-LD document to interface:", err)
	}

	framedDoc, err := proc.Frame(myInterface, frame, options) // do I need the options set in order to avoid the large context that seems to be generated?
	if err != nil {
		log.Println("Error when trying to frame document", err)
	}

	graph := framedDoc["@graph"]
	// ld.PrintDocument("JSON-LD graph section", graph)  // debug print....
	jsonm, err := json.MarshalIndent(graph, "", " ")
	if err != nil {
		log.Println("Error trying to marshal data", err)
	}

	dss := make([]OrgFrameRes, 0)
	err = json.Unmarshal(jsonm, &dss)
	if err != nil {
		log.Println("Error trying to unmarshal data to struct", err)
	}

	log.Printf("This is the dss:  %v\n", dss)
	return dss
}
