package millers

import (
	"fmt"
	"log"
	"time"

	"earthcube.org/Project418/gleaner/internal/millers/millerbleve"
	"earthcube.org/Project418/gleaner/internal/millers/millerfdpgraph"
	"earthcube.org/Project418/gleaner/internal/millers/millerfdptika"
	"earthcube.org/Project418/gleaner/internal/millers/millerfdptikajena"
	"earthcube.org/Project418/gleaner/internal/millers/millerprov"
	"earthcube.org/Project418/gleaner/internal/millers/millersgraph"
	"earthcube.org/Project418/gleaner/internal/millers/millershacl"
	"earthcube.org/Project418/gleaner/internal/millers/millersmock"
	"earthcube.org/Project418/gleaner/internal/millers/millerspatial"
	"earthcube.org/Project418/gleaner/internal/millers/millertika"
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
			millersmock.MockObjects(mc, as[d])
		}
	}

	if cs.Millers.Graph {
		for d := range as {
			millersgraph.GraphMillObjects(mc, as[d])
		}
	}

	if cs.Millers.Spatial {
		for d := range as {
			millerspatial.ProcessBucketObjects(mc, as[d])
		}
		// TODO add in saving the AOF file to the output directory
	}

	if cs.Millers.Shacl {
		for d := range as {
			millershacl.SHACLMillObjects(mc, as[d])
		}
		// TODO add in saving the AOF file to the output directory
	}

	if cs.Millers.Organic {
		for d := range as {
			millerbleve.GetObjects(mc, as[d])
		}
	}

	if cs.Millers.Prov {
		for d := range as {
			millerprov.MockObjects(mc, as[d], cs)
		}
	}

	if cs.Millers.Tika {
		for d := range as {
			millertika.TikaObjects(mc, as[d])
		}
	}

	if cs.Millers.FDPTika {
		for d := range as {
			millerfdptika.TikaObjects(mc, as[d])
		}
	}

	if cs.Millers.FDPTikaJena {
		for d := range as {
			millerfdptikajena.TikaObjects(mc, as[d])
		}
	}

	if cs.Millers.FDPGraph {
		for d := range as {
			millerfdpgraph.TikaObjects(mc, as[d])
		}
	}

	et := time.Now()
	diff := et.Sub(st)
	log.Printf("Miller end time: %s \n", et)
	log.Printf("Miller run time: %f \n", diff.Minutes())
}
