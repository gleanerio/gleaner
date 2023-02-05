package graph

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	configTypes "github.com/gleanerio/gleaner/internal/config"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
	"io"
	"strings"
	"sync"

	"github.com/gleanerio/gleaner/internal/common"
	minio "github.com/minio/minio-go/v7"
	"github.com/piprate/json-gold/ld"
	"github.com/spf13/viper"
)

// GraphNG is a new and improved RDF conversion
func GraphNG(mc *minio.Client, prefix string, v1 *viper.Viper) error {

	// read config file
	//miniocfg := v1.GetStringMapString("minio")
	//bucketName := miniocfg["bucket"] //   get the top level bucket for all of gleaner operations from config file
	bucketName, err := configTypes.GetBucketName(v1)
	// My go func controller vars
	semaphoreChan := make(chan struct{}, 10) // a blocking channel to keep concurrency under control (1 == single thread)
	defer close(semaphoreChan)
	wg := sync.WaitGroup{} // a wait group enables the main process a wait for goroutines to finish

	proc, options := common.JLDProc(v1) // Make a common proc and options to share with the upcoming go funcs

	// params for list objects calls
	// doneCh := make(chan struct{}) // , N) Create a done channel to control 'ListObjectsV2' go routine.
	// defer close(doneCh)           // Indicate to our routine to exit cleanly upon return.
	isRecursive := true

	// Need count for Spiffy progress line  (do I really want this?)
	// there has got to be a better way to get the count of objects in an object store
	oc := mc.ListObjects(context.Background(), bucketName, minio.ListObjectsOptions{Prefix: prefix, Recursive: isRecursive})
	// count := len(objectCh)
	var count int
	for x := range oc {
		count = count + 1
		if false {
			log.Println(x)
		}
	}

	// Old style bar is "ok" since done in sequence?
	bar := progressbar.Default(int64(count))
	objectCh := mc.ListObjects(context.Background(), bucketName, minio.ListObjectsOptions{Prefix: prefix, Recursive: isRecursive})
	// for object := range mc.ListObjects(context.Background(), bucketname, prefix, isRecursive, doneCh) {
	for object := range objectCh {
		wg.Add(1)
		go func(object minio.ObjectInfo) {
			semaphoreChan <- struct{}{}
			_, err := uploadObj2RDF(bucketName, "milled", mc, object, proc, options)
			if err != nil {
				log.Error("uploadObj2RDF", err) // need to log to an "errors" log file
			}

			wg.Done() // tell the wait group that we be done
			log.Debug("Doc:", object.Key, "error:", err)

			bar.Add(1) //bar1.Incr()
			<-semaphoreChan
		}(object)
	}
	wg.Wait()

	// TODO make a version of PipeCopy that generates Parquet version of graph
	// TODO..  then delete milled objects?
	log.Trace("Processed prefix:", prefix)
	millprefix := strings.ReplaceAll(prefix, "summoned", "milled")
	sp := strings.SplitAfterN(prefix, "/", 2)
	mcfg := v1.GetStringMapString("gleaner")

	rslt := fmt.Sprintf("results/%s/%s_graph.nq", mcfg["runid"], sp[1])
	log.Info("Assembling result graph for prefix:", prefix, "to:", millprefix)
	log.Info("Result graph will be at:", rslt)

	err = common.PipeCopyNG(rslt, bucketName, millprefix, mc)
	if err != nil {
		log.Error("Error on pipe copy:", err)
	} else {
		log.Info("Pipe copy for graph done")
	}

	return err
}

// func uploadObj2RDF(fo io.Reader, key string, mc *minio.Client) (string, int64, error) {
func uploadObj2RDF(bucketName, prefix string, mc *minio.Client, object minio.ObjectInfo, proc *ld.JsonLdProcessor, options *ld.JsonLdOptions) (string, error) {
	// object is an object reader
	stat, err := mc.StatObject(context.Background(), bucketName, object.Key, minio.GetObjectOptions{})
	if stat.Size > 100000 {
		log.Warn("retrieving a large object (", stat.Size, ") (this may be slow)", object.Key)
	}
	fo, err := mc.GetObject(context.Background(), bucketName, object.Key, minio.GetObjectOptions{})
	if err != nil {
		log.Error("minio.getObject", err)
		return "", err
	}

	key := object.Key // replace if new function idea works..

	var b bytes.Buffer
	bw := bufio.NewWriter(&b)

	_, err = io.Copy(bw, fo)
	if err != nil {
		log.Error("error copying:", err)
	}

	// TODO
	// Process the bytes in b to RDF (with randomized blank nodes)
	//log.Trace("JLD2NQ call")
	//rdf, err := common.JLD2nq(b.String(), proc, options)
	//if err != nil {
	//	return key, err
	//}
	//
	//log.Trace("blank node fix call")
	//rdfubn := GlobalUniqueBNodes(rdf)
	rdfubn, err := Obj2RDF(b.String(), proc, options)
	if err != nil {
		return key, err
	}
	milledkey := strings.ReplaceAll(key, ".jsonld", ".rdf")
	milledkey = strings.ReplaceAll(milledkey, "summoned/", "")

	// make an object with prefix like  scienceorg-dg/objectname.rdf  (where is had .jsonld before)
	objectName := fmt.Sprintf("%s/%s", prefix, milledkey)
	usermeta := make(map[string]string) // what do I want to know?
	usermeta["origfile"] = key

	// Upload the file
	_, err = LoadToMinio(rdfubn, bucketName, objectName, mc)
	if err != nil {
		return objectName, err
	}

	return objectName, nil
}

func Obj2RDF(jsonld string, proc *ld.JsonLdProcessor, options *ld.JsonLdOptions) (string, error) {

	// Process the bytes in b to RDF (with randomized blank nodes)
	log.Trace("JLD2NQ call")
	rdf, err := common.JLD2nq(jsonld, proc, options)
	if err != nil {
		return "", err
	}

	log.Trace("blank node fix call")
	rdfubn := GlobalUniqueBNodes(rdf)

	return rdfubn, nil
}

// sugar function for the ui bar
func rightPad2Len(s string, padStr string, overallLen int) string {
	var padCountInt int
	padCountInt = 1 + ((overallLen - len(padStr)) / len(padStr))
	var retStr = s + strings.Repeat(padStr, padCountInt)
	return retStr[:overallLen]
}
