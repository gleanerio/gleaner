package millers

import (
	"fmt"
	"log"
	"time"

	"earthcube.org/Project418/gleaner/millers/millerbleve"
	"earthcube.org/Project418/gleaner/millers/millerfdpgraph"
	"earthcube.org/Project418/gleaner/millers/millerfdptika"
	"earthcube.org/Project418/gleaner/millers/millerprov"
	"earthcube.org/Project418/gleaner/millers/millersgraph"
	"earthcube.org/Project418/gleaner/millers/millersmock"
	"earthcube.org/Project418/gleaner/millers/millerspatial"
	"earthcube.org/Project418/gleaner/millers/millertika"
	"earthcube.org/Project418/gleaner/millers/millerutils"
	"earthcube.org/Project418/gleaner/utils"
)

func Millers(cs utils.Config, rundir string) {

	millerutils.RunDir = rundir // set output dir for graph, fdpgraph and prov
	// set output for bleve and fdptika and tika

	st := time.Now()
	log.Printf("Miller start time: %s \n", st) // Log the time at start for the record

	mc := utils.MinioConnection(cs) // minio connection

	buckets, err := utils.ListBuckets(mc)
	if err != nil {
		log.Println(err)
	}

	fmt.Println("Bucket list...")
	for _, bucket := range buckets {
		log.Println(bucket.Name) // for no real reason.. :)
	}

	// Make an array "as" of active buckets to process...
	as := []string{}
	for i := range cs.Sources {
		if cs.Sources[i].Active == true {
			as = append(as, cs.Sources[i].Name)
		}
	}

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

	if cs.Millers.Organic {
		for d := range as {
			millerbleve.GetObjects(mc, as[d])
		}
	}

	if cs.Millers.Prov {
		for d := range as {
			millerprov.MockObjects(mc, as[d])
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
