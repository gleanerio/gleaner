package common

import (
	"bytes"
	"log"

	"context"
	minio "github.com/minio/minio-go/v7"
)

// Entry is a struct holding the json-ld metadata and data (the text)
type Entry struct {
	Bucketname string
	Key        string
	Urlval     string
	Sha1val    string
	Jld        string
}

// GetMillObjects gets the objects for a given miller
func GetMillObjects(mc *minio.Client, prefix string) []Entry {
	doneCh := make(chan struct{}) // Create a done channel to control 'ListObjectsV2' go routine.
	defer close(doneCh)           // Indicate to our routine to exit cleanly upon return.
	isRecursive := true
	bucketname := "gleaner-summoned"
	opts := minio.ListObjectsOptions{
		Recursive: isRecursive,
		Prefix:    prefix,
	}
	//objectCh := mc.ListObjects(context.Background(), bucketname, prefix, isRecursive, doneCh)
	objectCh := mc.ListObjects(context.Background(), bucketname, opts)

	var entries []Entry

	for object := range objectCh {
		if object.Err != nil {
			log.Println(object.Err)
			return nil
		}

		fo, err := mc.GetObject(context.Background(), bucketname, object.Key, minio.GetObjectOptions{})
		if err != nil {
			log.Println(err)
			return nil
		}

		oi, err := fo.Stat()
		if err != nil {
			log.Println("Issue with reading an object..  should I just fatal on this to make sure?")
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

	log.Println(len(entries))
	// multiCall(entries)

	return entries
}
