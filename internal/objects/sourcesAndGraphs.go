package objects

import (
	"log"

	"github.com/earthcubearchitecture-project418/gleaner/internal/summoner/acquire"
)

// Return all sources and sitegraph domains
func SourcesAndGraphs() []acquire.Sources {
	var domains []acquire.Sources
	var sm []acquire.Sources

	err := v1.UnmarshalKey("sources", &sm)
	if err != nil {
		log.Println(err)
	}

	var sg []acquire.Sources

	err = v1.UnmarshalKey("sitegraphs", &sg)
	if err != nil {
		log.Println(err)
	}

	domains = append(sg, sm...)

	return domains

}
