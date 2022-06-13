package graph

import (
	"bytes"
	"context"
	log "github.com/sirupsen/logrus"

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
		log.Error(bucketName, "/", objectName, "error", err)
		// TODO   should return 0, err here and deal with it on the other end
	}

	log.Trace("Uploaded Bucket:", bucketName, "File:", objectName, "Size", n)

	return n.Size, nil
}
