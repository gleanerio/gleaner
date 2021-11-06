package shapes

import (
	"context"
	"fmt"
	configTypes "github.com/gleanerio/gleaner/internal/config"
	"github.com/gosuri/uiprogress"
	log "github.com/sirupsen/logrus"
	"strings"
	"sync"

	"github.com/gleanerio/gleaner/internal/common"
	minio "github.com/minio/minio-go/v7"
	"github.com/spf13/viper"
)

// ShapeNG is a new and improved RDF conversion
func ShapeNG(mc *minio.Client, prefix string, v1 *viper.Viper) error {

	// read config file
	//miniocfg := v1.GetStringMapString("minio")
	//bucketName := miniocfg["bucket"] //   get the top level bucket for all of gleaner operations from config file
	bucketName, err := configTypes.GetBucketName(v1)

	loadShapeFiles(mc, v1) // TODO, this should be done in main

	// My go func controller vars
	semaphoreChan := make(chan struct{}, 30) // a blocking channel to keep concurrency under control (1 == single thread)
	defer close(semaphoreChan)
	wg := sync.WaitGroup{} // a wait group enables the main process a wait for goroutines to finish

	// Make a common proc and options to share with the upcoming go funcs
	proc, options := common.JLDProc(v1)

	// params for list objects calls
	doneCh := make(chan struct{}) // , N) Create a done channel to control 'ListObjectsV2' go routine.
	defer close(doneCh)           // Indicate to our routine to exit cleanly upon return.
	isRecursive := true

	// Spiffy progress line
	uiprogress.Start()
	x := 0 // ugh..  why won't len(oc) work..   buffered channel issue I assume?
	opts := minio.ListObjectsOptions{
		Recursive: isRecursive,
		Prefix:    prefix,
	}
	//for range mc.ListObjectsV2(bucketname, prefix, isRecursive, doneCh)
	for range mc.ListObjects(context.Background(), bucketName, opts) {
		x = x + 1
	}
	count := x
	bar3 := uiprogress.AddBar(count).PrependElapsed().AppendCompleted()
	bar3.PrependFunc(func(b *uiprogress.Bar) string {
		return rightPad2Len("shacl", " ", 15)
	})
	bar3.Fill = '-'
	bar3.Head = '>'
	bar3.Empty = ' '

	// TODO get the list of shape files in the shape bucket
	//for shape := range mc.ListObjectsV2(bucketname, "shapes", isRecursive, doneCh) {
	opts2 := minio.ListObjectsOptions{
		Recursive: isRecursive,
		Prefix:    "shapes",
	}
	for shape := range mc.ListObjects(context.Background(), bucketName, opts2) {
		// log.Printf("Checking data graphs against shape graph: %s\n", m[j])

		//for object := range mc.ListObjectsV2(bucketname, prefix, isRecursive, doneCh) {
		for object := range mc.ListObjects(context.Background(), prefix, minio.ListObjectsOptions{
			Recursive: isRecursive,
		}) {
			wg.Add(1)
			go func(object minio.ObjectInfo) {
				semaphoreChan <- struct{}{}
				//status := shaclTest(e[k].Urlval, e[k].Jld, m[j].Key, m[j].Jld, &gb)
				_, err := shaclTestNG(v1, bucketName, "verified", mc, object, shape, proc, options)
				if err != nil {
					log.Println(err) // need to log to an "errors" log file
				}

				// _, err := obj2RDF(bucketName, prefix, mc, object, proc, options)

				wg.Done() // tell the wait group that we be done
				// log.Printf("Doc: %s error: %v ", name, err) // why print the status??

				bar3.Incr()
				<-semaphoreChan
			}(object)
		}
	}
	wg.Wait()

	// uiprogress.Stop()

	// // all done..  write the full graph to the object store
	// log.Printf("Saving full graph to  gleaner milled:  Ref: %s/%s", bucketName, prefix)

	// //pipeCopyNG(mcfg["runid"], "gleaner-milled", fmt.Sprintf("%s-sg", prefix), mc)
	// // TODO fix this with correct variables
	// pipeCopyNG(mcfg["runid"], "gleaner-milled", fmt.Sprintf("%s-sg", prefix), mc)
	// log.Printf("Saving datagraph to:  %s/%s", bucketName, prefix)

	// log.Printf("Processed prefix: %s", prefix)
	millprefix := strings.ReplaceAll(prefix, "summoned", "verified")
	sp := strings.SplitAfterN(prefix, "/", 2)
	mcfg := v1.GetStringMapString("gleaner")
	rslt := fmt.Sprintf("results/%s/%s_verified.nq", mcfg["runid"], sp[1])
	log.Printf("Assembling result graph for prefix: %s to: %s", prefix, millprefix)
	log.Printf("Result graph will be at: %s", rslt)

	err = common.PipeCopyNG(rslt, bucketName, millprefix, mc)
	if err != nil {
		log.Printf("Error on pipe copy: %s", err)
	} else {
		log.Println("Pipe copy for shacl done")
	}

	return err
}

//  ---------- func below are dupes..  they will be moved to a commons

// sugar function for the ui bar
func rightPad2Len(s string, padStr string, overallLen int) string {
	var padCountInt int
	padCountInt = 1 + ((overallLen - len(padStr)) / len(padStr))
	var retStr = s + strings.Repeat(padStr, padCountInt)
	return retStr[:overallLen]
}

// func pipeCopyNG(name, bucket, prefix string, mc *minio.Client) error {
// 	log.Println("Start pipe reader / writer sequence")

// 	pr, pw := io.Pipe()     // TeeReader of use?
// 	lwg := sync.WaitGroup{} // work group for the pipe writes...
// 	lwg.Add(2)

// 	// params for list objects calls
// 	doneCh := make(chan struct{}) // , N) Create a done channel to control 'ListObjectsV2' go routine.
// 	defer close(doneCh)           // Indicate to our routine to exit cleanly upon return.
// 	isRecursive := true

// 	go func() {
// 		defer lwg.Done()
// 		defer pw.Close()
// 		for object := range mc.ListObjectsV2(bucket, prefix, isRecursive, doneCh) {
// 			fo, err := mc.GetObject(bucket, object.Key, minio.GetObjectOptions{})
// 			if err != nil {
// 				fmt.Println(err)
// 			}

// 			var b bytes.Buffer
// 			bw := bufio.NewWriter(&b)

// 			_, err = io.Copy(bw, fo)
// 			if err != nil {
// 				log.Println(err)
// 			}

// 			pw.Write(b.Bytes())
// 		}

// 	}()

// 	// go function to write to minio from pipe
// 	go func() {
// 		defer lwg.Done()
// 		_, err := mc.PutObject(bucketName, name, pr, -1, minio.PutObjectOptions{})
// 		if err != nil {
// 			log.Println(err)
// 		}
// 	}()

// 	// Note: We can also make a file and pipe write to that, keep this code around in case
// 	// f, err := os.Create(fmt.Sprintf("%s_graph.nq", prefix))  // needs a f.Close() later
// 	// if err != nil {
// 	// 	log.Println(err)
// 	// }
// 	// go function to write to file from pipe
// 	// go func() {
// 	// 	defer lwg.Done()
// 	// 	if _, err := io.Copy(f, pr); err != nil {
// 	// 		log.Fatal(err)
// 	// 	}
// 	// }()

// 	lwg.Wait() // wait for the pipe read writes to finish
// 	pw.Close()
// 	pr.Close()

// 	return nil
// }
