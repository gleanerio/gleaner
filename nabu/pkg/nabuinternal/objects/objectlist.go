package objects

import (
	"context"
	"github.com/gleanerio/gleaner/nabu/pkg/config"
	"github.com/minio/minio-go/v7"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"sync"
)

func ObjectList(v1 *viper.Viper, mc *minio.Client, prefix string) ([]string, error) {
	//objs := v1.GetStringMapString("objects")
	//objs,_ := config.GetObjectsConfig(v1)
	bucketName, _ := config.GetBucketName(v1)

	// My go func controller vars
	semaphoreChan := make(chan struct{}, 1) // a blocking channel to keep concurrency under control (1 == single thread)
	defer close(semaphoreChan)
	wg := sync.WaitGroup{} // a wait group enables the main process a wait for goroutines to finish

	// params for list objects calls
	doneCh := make(chan struct{}) // , N) Create a done channel to control 'ListObjectsV2' go routine.
	defer close(doneCh)           // Indicate to our routine to exit cleanly upon return.
	// isRecursive := true

	oa := []string{}

	// for object := range mc.ListObjectsV2(objs["bucket"], objs["prefix"], isRecursive, doneCh) {
	// for object := range mc.ListObjects(context.Background(), objs["bucket"],
	for object := range mc.ListObjects(context.Background(), bucketName,
		minio.ListObjectsOptions{Prefix: prefix, Recursive: true}) {

		wg.Add(1)
		go func(object minio.ObjectInfo) {
			oa = append(oa, object.Key) // WARNING  append is not always thread safe..   wg of 1 till I address this
			wg.Done()                   // tell the wait group that we be done
			<-semaphoreChan
		}(object)
		wg.Wait()
	}

	// log.Printf("%s:%s object count: %d\n", objs["bucket"], prefix, len(oa))
	log.Printf("%s:%s object count: %d\n", bucketName, prefix, len(oa))
	return oa, nil
}
