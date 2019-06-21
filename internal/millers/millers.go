package millers

import (
	"fmt"
	"log"
	"time"

	"earthcube.org/Project418/gleaner/internal/millers/fdpgraph"
	"earthcube.org/Project418/gleaner/internal/millers/fdptika"
	"earthcube.org/Project418/gleaner/internal/millers/fdptikajena"
	"earthcube.org/Project418/gleaner/internal/millers/graph"
	"earthcube.org/Project418/gleaner/internal/millers/mock"
	"earthcube.org/Project418/gleaner/internal/millers/ner"
	"earthcube.org/Project418/gleaner/internal/millers/prov"
	"earthcube.org/Project418/gleaner/internal/millers/shapes"
	"earthcube.org/Project418/gleaner/internal/millers/spatial"
	"earthcube.org/Project418/gleaner/internal/millers/textindex"
	"earthcube.org/Project418/gleaner/internal/millers/tika"
	"earthcube.org/Project418/gleaner/pkg/utils"
	"github.com/minio/minio-go"
)

// func Millers(cs utils.Config, rundir string) {
func Millers(mc *minio.Client, cs utils.Config) {

	// millerutils.RunDir = rundir // set output dir for graph, fdpgraph and prov
	// set output for bleve and fdptika and tika

	st := time.Now()
	log.Printf("Miller start time: %s \n", st) // Log the time at start for the record

	// mc := utils.MinioConnection(cs) // minio connection

	// Get and print the bucket list for no reason what so ever.....
	buckets, err := utils.ListBuckets(mc)
	if err != nil {
		log.Println(err)
	}

	fmt.Println("Bucket list...")
	for _, bucket := range buckets {
		log.Println(bucket.Name)
	}

	// Make an array "as" of active buckets to process...
	as := []string{}
	for i := range cs.Sources {
		if cs.Sources[i].Active == true {
			as = append(as, cs.Sources[i].Name)
		}
	}

	// TODO easy concurency hidden here!!!!
	// Start calling the millers

	// Mock is just a template miller..  prints resource entries only...
	if cs.Millers.Mock {
		for d := range as {
			mock.MockObjects(mc, as[d])
		}
	}

	if cs.Millers.Graph {
		for d := range as {
			graph.GraphMillObjects(mc, as[d], cs)
		}
	}

	if cs.Millers.Spatial {
		for d := range as {
			spatial.ProcessBucketObjects(mc, as[d])
		}
		// TODO add in saving the AOF file to the output directory
	}

	if cs.Millers.Shacl {
		for d := range as {
			shapes.SHACLMillObjects(mc, as[d], cs)
		}
		// TODO add in saving the AOF file to the output directory
	}

	if cs.Millers.Organic {
		for d := range as {
			textindex.GetObjects(mc, as[d])
		}
	}

	if cs.Millers.Prov {
		for d := range as {
			prov.MockObjects(mc, as[d], cs)
		}
	}

	if cs.Millers.Tika {
		for d := range as {
			tika.TikaObjects(mc, as[d])
		}
	}

	if cs.Millers.FDPTika {
		for d := range as {
			fdptika.TikaObjects(mc, as[d], cs)
		}
	}

	if cs.Millers.FDPTikaJena {
		for d := range as {
			fdptikajena.TikaObjects(mc, as[d], cs)
		}
	}

	if cs.Millers.FDPGraph {
		for d := range as {
			fdpgraph.TikaObjects(mc, as[d], cs)
		}
	}

	if cs.Millers.NER {
		for d := range as {
			ner.NERObjects(mc, as[d])
		}
	}

	et := time.Now()
	diff := et.Sub(st)
	log.Printf("Miller end time: %s \n", et)
	log.Printf("Miller run time: %f \n", diff.Minutes())
}
