package graph

import (
	"bytes"
	"context"
	"log"

	minio "github.com/minio/minio-go/v7"
)

// LoadToMinio loads jsonld into the specified bucket
func LoadToMinio(jsonld, bucketName, objectName string, mc *minio.Client) (int64, error) {

	// set up some elements for PutObject
	contentType := "application/ld+json"
	b := bytes.NewBufferString(jsonld)
	usermeta := make(map[string]string) // what do I want to know?
	// usermeta["url"] = urlloc
	// usermeta["sha1"] = bss

	//log.Println(bucketName)
	// Upload the zip file with FPutObject
	n, err := mc.PutObject(context.Background(), bucketName, objectName, b, int64(b.Len()), minio.PutObjectOptions{ContentType: contentType, UserMetadata: usermeta})
	if err != nil {
		log.Printf("%s/%s", bucketName, objectName)
		log.Println(err)
		// TODO   should return 0, err here and deal with it on the other end
	}

	// log.Printf("#%d Uploaded Bucket:%s File:%s Size %d\n", i, bucketName, objectName, n)

	return n.Size, nil
}
