package millers

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"

	"github.com/gleanerio/gleaner/internal/config"
	"github.com/gleanerio/gleaner/internal/millers/graph"
	"github.com/minio/minio-go/v7"
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

	// Put the sources in the config file into a struct
	//var domains []Sources
	//err := v1.UnmarshalKey("sources", &domains)

	//domains, err := config.GetSources(v1)
	domains, err := config.GetActiveSources(v1)
	if err != nil {
		log.Println(err)
	}

	// Make an array "as" of active buckets to process...
	as := []string{}
	for i := range domains {
		m := fmt.Sprintf("summoned/%s", domains[i].Name)
		as = append(as, m)
		log.Printf("Adding bucket to milling list: %s\n", m)
	}

	// Make array of prov buckets..  sad I have to do this..  I could just pass
	// the domains and let each miller now pick where to get things from.  I
	// only had to add this due to the prov data not being in summoned

	ap := []string{}
	for i := range domains {
		m := fmt.Sprintf("prov/%s", domains[i].Name)
		ap = append(ap, m)
		log.Printf("Adding bucket to prov building list: %s\n", m)
	}

	mcfg := v1.GetStringMapString("millers") // get the millers we want to run from the config file

	// Graph is the miller to convert from JSON-LD to nquads with validation of well formed
	// TODO  none of these (graph, shacl, prov) deal with the returned error
	if mcfg["graph"] == "true" {
		for d := range as {
			graph.GraphNG(mc, as[d], v1)
		}
	}

	// if mcfg["shacl"] == "true" {
	// 	for d := range as {
	// 		shapes.ShapeNG(mc, as[d], v1)
	// 		// shapes.SHACLMillObjects(mc, as[d], v1)
	// 	}
	// }

	// if mcfg["prov"] == "true" {
	// 	for d := range ap {
	// 		graph.AssembleObjs(mc, ap[d], v1)
	// 	}
	// }

	// Time report
	et := time.Now()
	diff := et.Sub(st)
	log.Printf("Miller end time: %s \n", et)
	log.Printf("Miller run time: %f \n", diff.Minutes())
}
