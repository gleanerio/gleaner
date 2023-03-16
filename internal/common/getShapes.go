package common

import (
	"bytes"
	"context"
	minio "github.com/minio/minio-go/v7"
	log "github.com/sirupsen/logrus"
)

// GetShapeGraphs gets the shape graphs the shacl miller will use.
// Currently this is a basic duplication of GetObjects but later this may
// pull shapes from geoschemas.org so this will change completely.  For
// it pulls from the object store
func GetShapeGraphs(mc *minio.Client, bucketname string) []Entry {
	doneCh := make(chan struct{}) // Create a done channel to control 'ListObjectsV2' go routine.
	defer close(doneCh)           // Indicate to our routine to exit cleanly upon return.
	isRecursive := true
	opts := minio.ListObjectsOptions{
		Recursive: isRecursive,
		//Prefix:    "my-prefixname",
	}
	//objectCh := mc.ListObjects(bucketname, "", isRecursive, doneCh)
	objectCh := mc.ListObjects(context.Background(), bucketname, opts)
	var entries []Entry

	for object := range objectCh {
		if object.Err != nil {
			log.Error(object.Err)
			return nil
		}

		//fo, err := mc.GetObject(bucketname, object.Key, minio.GetObjectOptions{})
		fo, err := mc.GetObject(context.Background(), bucketname, object.Key, minio.GetObjectOptions{})
		if err != nil {
			log.Error(err)
			return nil
		}

		oi, err := fo.Stat()
		if err != nil {
			log.Error("Issue with reading an object..  should I just fatal on this to make sure?", err)
		}
		urlval := ""
		sha1val := ""
		if len(oi.Metadata["X-Amz-Meta-Url"]) > 0 {
			urlval = oi.Metadata["X-Amz-Meta-Url"][0] // also have  X-Amz-Meta-Sha1
		}
		if len(oi.Metadata["X-Amz-Meta-Sha1"]) > 0 {
			sha1val = oi.Metadata["X-Amz-Meta-Sha1"][0]
		}

		buf := new(bytes.Buffer)
		buf.ReadFrom(fo)
		jld := buf.String() // Does a complete copy of the bytes in the buffer.

		// Mock call for some validation (and a template for other millers)
		// Mock(bucketname, object.Key, urlval, sha1val, jld)
		entry := Entry{Bucketname: bucketname, Key: object.Key, Urlval: urlval, Sha1val: sha1val, Jld: jld}
		entries = append(entries, entry)

	}

	log.Debug(len(entries))
	// multiCall(entries)

	return entries
}
