package objects

import (
	"bytes"
	"context"
	log "github.com/sirupsen/logrus"

	"github.com/minio/minio-go/v7"
)

// GetS3Bytes simply pulls the byes of an object from the store
func GetS3Bytes(mc *minio.Client, bucket, object string) ([]byte, string, error) {
	fo, err := mc.GetObject(context.Background(), bucket, object, minio.GetObjectOptions{})
	if err != nil {
		log.Info(err)
		return nil, "", err
	}
	defer fo.Close()

	oi, err := fo.Stat()
	if err != nil {
		log.Infof("Issue with reading an object:  %s%s", bucket, object)
	}

	// resuri := ""

	// urlval := ""
	// if len(oi.Metadata["X-Amz-Meta-Url"]) > 0 {
	// 	urlval = oi.Metadata["X-Amz-Meta-Url"][0] // also have  X-Amz-Meta-Sha1
	// }

	// shaval := ""
	// if len(oi.Metadata["X-Amz-Meta-Sha256"]) > 0 {
	// 	shaval = oi.Metadata["X-Amz-Meta-Sha256"][0]
	// }

	dgraph := ""
	if len(oi.Metadata["X-Amz-Meta-Dgraph"]) > 0 {
		dgraph = oi.Metadata["X-Amz-Meta-Dgraph"][0]
	}

	// fmt.Printf("%s %s %s \n", urlval, sha1val, resuri)

	// TODO  set an upper byte size  limit here and return error if the size is too big
	// TODO  why was this done, return size too and let the calling function worry about it...????
	//sz := oi.Size        // what type is this...
	//if sz > 1073741824 { // if bigger than 1 GB (which is very small) move on
	//	return nil, "", errors.New("gets3bytes says file above processing size threshold")
	//}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(fo)
	if err != nil {
		return nil, "", err
	}

	bb := buf.Bytes() // Does a complete copy of the bytes in the buffer.

	return bb, dgraph, err
}
