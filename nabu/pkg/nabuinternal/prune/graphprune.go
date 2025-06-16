package prune

import (
	"fmt"

	"github.com/gleanerio/gleaner/nabu/pkg/config"
	"github.com/gleanerio/gleaner/nabu/pkg/nabuinternal/graph"
	"github.com/gleanerio/gleaner/nabu/pkg/nabuinternal/objects"
	"github.com/minio/minio-go/v7"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Snip removes graphs in TS not in object store
func Snip(v1 *viper.Viper, mc *minio.Client) error {
	var pa []string
	//err := v1.UnmarshalKey("objects.prefix", &pa)

	objs, err := config.GetObjectsConfig(v1)
	bucketName, _ := config.GetBucketName(v1)
	if err != nil {
		log.Error(err)
	}
	pa = objs.Prefix

	for p := range pa {
		// collect the objects associated with the source
		oa, err := objects.ObjectList(v1, mc, pa[p])
		if err != nil {
			log.Error(err)
			return err
		}

		// collect the named graphs from graph associated with the source
		ga, err := graphList(v1, pa[p])
		if err != nil {
			log.Error(err)
			return err
		}

		// convert the object names to the URN pattern used in the graph
		// and make a map where key = URN, value = object name
		// NOTE:  since later we want to look up the object based the URN
		// we will do it this way since mapswnat you to know a key, not a value, when
		// querying them.
		// This is OK since all KV pairs involve unique keys and unique values
		var oam = map[string]string{}
		for x := range oa {
			g, err := graph.MakeURN(v1, oa[x])
			if err != nil {
				log.Error("MakeURN error: %v\n", err)
			}
			oam[g] = oa[x] // key (URN)= value (object prefixpath)
		}

		// make an array of just the values for use with findMissing and difference functions
		// we have in this package
		var oag []string // array of all keys
		for k, _ := range oam {
			oag = append(oag, k)
		}

		//compare lists, anything IN graph not in objects list should be removed
		d := difference(ga, oag)  // return items in ga that are NOT in oag, we should remove these
		m := findMissing(oag, ga) // return items from oag we need to add

		fmt.Printf("Current graph items: %d  Cuurent object items: %d\n", len(ga), len(oag))
		fmt.Printf("Orphaned items to remove: %d\n", len(d))
		fmt.Printf("Missing items to add: %d\n", len(m))

		log.WithFields(log.Fields{"prefix": pa[p], "graph items": len(ga), "object items": len(oag), "difference": len(d),
			"missing": len(m)}).Info("Nabu Prune")

		// For each in d will delete that graph
		if len(d) > 0 {
			bar := progressbar.Default(int64(len(d)))
			for x := range d {
				log.Infof("Removed graph: %s\n", d[x])
				_, err = graph.Drop(v1, d[x])
				if err != nil {
					log.Error("Progress bar update issue: %v\n", err)
				}
				err = bar.Add(1)
				if err != nil {
					log.Error("Progress bar update issue: %v\n", err)
				}
			}
		}

		ep := v1.GetString("flags.endpoint")
		spql, err := config.GetEndpoint(v1, ep, "bulk")
		if err != nil {
			log.Error(err)
		}

		//// load new ones
		//spql, err := config.GetSparqlConfig(v1)
		//if err != nil {
		//	log.Error("prune -> config.GetSparqlConfig %v\n", err)
		//}

		if len(m) > 0 {
			bar2 := progressbar.Default(int64(len(m)))
			log.Info("uploading missing %n objects", len(m))
			for x := range m {
				np := oam[m[x]]
				log.Tracef("Add graph: %s  %s \n", m[x], np)
				_, err := objects.PipeLoad(v1, mc, bucketName, np, spql.URL)
				if err != nil {
					log.Error("prune -> pipeLoad %v\n", err)
				}
				err = bar2.Add(1)
				if err != nil {
					log.Error("Progress bar update issue: %v\n", err)
				}
			}
		}
	}

	return nil
}
