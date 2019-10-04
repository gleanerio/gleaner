package millers

import (
	"fmt"
	"log"
	"time"

	"earthcube.org/Project418/gleaner/internal/millers/graph"
	"earthcube.org/Project418/gleaner/internal/millers/prov"
	"earthcube.org/Project418/gleaner/internal/millers/shapes"
	"github.com/minio/minio-go"
	"github.com/spf13/viper"
)

type Sources struct {
	Name     string
	Logo     string
	URL      string
	Headless bool
	// SitemapFormat string
	// Active        bool
}

// Millers is our main controller for calling the various milling paths we will
// do on the JSON-LD data graphs
func Millers(mc *minio.Client, v1 *viper.Viper) {
	st := time.Now()
	log.Printf("Miller start time: %s \n", st) // Log the time at start for the record

	var domains []Sources
	err := v1.UnmarshalKey("sources", &domains)
	if err != nil {
		log.Println(err)
	}

	// Make an array "as" of active buckets to process...
	as := []string{}
	for i := range domains {
		m := fmt.Sprintf("%s", domains[i].Name)
		as = append(as, m)
		log.Printf("Adding bucket to milling list: %s\n", m)
	}

	mcfg := v1.GetStringMapString("millers")

	// Graph is the miller to convert from JSON-LD to nquads with validation of well formed
	if mcfg["graph"] == "true" {
		graph.MillerSetup(mc, as, v1) // kv based function (disk based with memory mapping)
		//for d := range as {
		// graph.MillObjects(mc, as[d], cs)  // old memory based function
		// TODO really each of these can be a go func.call .
		// be sure to update the file name  (buckets can stay the same since different files)
		//	graph.Miller(mc, as[d], cs) // kv based function (disk based with memory mapping)
		//}
	}

	if mcfg["shacl"] == "true" {
		for d := range as {
			shapes.SHACLMillObjects(mc, as[d], v1)
		}
	}

	if mcfg["prov"] == "true" {
		for d := range as {
			prov.MockObjects(mc, as[d], v1)
		}
	}

	et := time.Now()
	diff := et.Sub(st)
	log.Printf("Miller end time: %s \n", et)
	log.Printf("Miller run time: %f \n", diff.Minutes())
}
