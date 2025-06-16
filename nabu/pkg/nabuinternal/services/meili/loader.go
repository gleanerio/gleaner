package meili

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gleanerio/gleaner/nabu/pkg/config"
	"github.com/gleanerio/gleaner/nabu/pkg/nabuinternal/objects"

	"github.com/meilisearch/meilisearch-go"
	log "github.com/sirupsen/logrus"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/spf13/viper"
)

// ObjectAssembly collects the objects from a bucket to load
func ObjectAssembly(v1 *viper.Viper, mc *minio.Client) error {
	fmt.Println("MeiliSearch index started, reading server from GLEANERIO_MEILI_SERVER")
	host := getEnv("GLEANERIO_MEILI_SERVER")

	bucketName, _ := config.GetBucketName(v1)
	objCfg, _ := config.GetObjectsConfig(v1)
	pa := objCfg.Prefix

	msc := meilisearch.NewClient(meilisearch.ClientConfig{
		Host: host,
	})

	var err error

	// TODO  even though this is removed, it would be good proactive to give a
	// unique name based on something like time and other parameters
	name := "bulkobject.json"

	for p := range pa {
		log.Printf("ToJSONArray for %s", pa[p])
		ToJSONArray(name, bucketName, pa[p], mc)
		if err != nil {
			return err
		}
	}

	for p := range pa {
		log.Printf("dofunc for %s  %s", p, fmt.Sprintf("%s/%s", pa[p], name))
		// will need a function call at some point to work with the new object
		r, err := docfunc(v1, mc, msc, bucketName, fmt.Sprintf("%s/%s", pa[p], name), "endpoint")
		if err != nil {
			log.Println(err)
		}
		log.Printf("Return from docfunc: %s", string(r))
	}

	// TODO  remove the temporary object?
	for p := range pa {
		log.Printf("remove %s for %s", name, pa[p])
		opts := minio.RemoveObjectOptions{}
		err = mc.RemoveObject(context.Background(), bucketName, fmt.Sprintf("%s/%s", pa[p], name), opts)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	return err
}

func getEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = "default"
	}
	return value
}

func docfunc(v1 *viper.Viper, mc *minio.Client, msc *meilisearch.Client, bucketName string, item string, endpoint string) ([]byte, error) {
	// get item
	b, _, err := objects.GetS3Bytes(mc, bucketName, item)
	if err != nil {
		return nil, err
	}

	var doc interface{} // why was this a map[string]interface{} before?  bulk vs single..  doesn't seem so..
	//json.Unmarshal([]byte(value), &doc)
	json.Unmarshal(b, &doc)

	r, err := msc.Index("iow").AddDocuments(doc, "id")
	if err != nil {
		log.Println(err)
	}

	log.Println(r)

	return nil, nil
}
