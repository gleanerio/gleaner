package graph

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/earthcubearchitecture-project418/gleaner/internal/common"
	"github.com/gosuri/uiprogress"

	minio "github.com/minio/minio-go"
	"github.com/spf13/viper"
)

// Source struct
type Sources struct {
	Name     string
	Logo     string
	URL      string
	Headless bool
	// SitemapFormat string
	// Active        bool
}

// AssembleObjs builds a prov graph from the generated proj JSON-LD files
func AssembleObjs(mc *minio.Client, prefix string, v1 *viper.Viper) error {
	bucketname := "gleaner" //  "gleaner-summoned"
	// prefix := fmt.Sprintf("summoned/%s", suffix)

	// My go func controller vars
	semaphoreChan := make(chan struct{}, 10) // a blocking channel to keep concurrency under control (1 == single thread)
	defer close(semaphoreChan)
	wg := sync.WaitGroup{} // a wait group enables the main process a wait for goroutines to finish

	// Make a common proc and options to share with the upcoming go funcs
	proc, options := common.JLDProc(v1)

	// params for list objects calls
	doneCh := make(chan struct{}) // , N) Create a done channel to control 'ListObjectsV2' go routine.
	defer close(doneCh)           // Indicate to our routine to exit cleanly upon return.
	isRecursive := true

	// Spiffy progress line  (do I really want this?)
	uiprogress.Start()
	x := 0 // ugh..  why won't len(oc) work..   buffered channel issue I assume?
	for range mc.ListObjectsV2(bucketname, prefix, isRecursive, doneCh) {
		x = x + 1
	}
	count := x
	bar1 := uiprogress.AddBar(count).PrependElapsed().AppendCompleted()
	bar1.PrependFunc(func(b *uiprogress.Bar) string {
		return rightPad2Len("miller", " ", 15)
	})
	bar1.Fill = '-'
	bar1.Head = '>'
	bar1.Empty = ' '

	for object := range mc.ListObjectsV2(bucketname, prefix, isRecursive, doneCh) {
		wg.Add(1)
		go func(object minio.ObjectInfo) {
			semaphoreChan <- struct{}{}
			_, err := obj2RDF(bucketname, "milled", mc, object, proc, options)
			if err != nil {
				log.Println(err) // need to log to an "errors" log file
			}

			wg.Done() // tell the wait group that we be done
			// log.Printf("Doc: %s error: %v ", name, err) // why print the status??

			bar1.Incr()
			<-semaphoreChan
		}(object)
	}
	wg.Wait()

	// uiprogress.Stop()

	// all done..  write the full graph to the object store
	// log.Printf("Building result graph from: %s/milled-dg/%s", bucketname, prefix)

	// log.Printf("Processed prefix: %s", prefix)
	millprefix := strings.ReplaceAll(prefix, "prov", "milled/prov")
	sp := strings.SplitAfterN(prefix, "/", 2)
	mcfg := v1.GetStringMapString("gleaner")
	rslt := fmt.Sprintf("results/%s/%s_prov.nq", mcfg["runid"], sp[1])
	log.Printf("Assembling prov graph for prefix: %s to: %s", prefix, millprefix)
	log.Printf("Prov graph will be at: %s", rslt)

	err := common.PipeCopyNG(rslt, "gleaner", millprefix, mc)
	if err != nil {
		log.Printf("Error on pipe copy: %s", err)
	} else {
		log.Println("Pipe copy for graph done")
	}

	return err
}
