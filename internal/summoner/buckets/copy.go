package buckets

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	minio "github.com/minio/minio-go/v7"
)

type BucketObjects struct {
	Name string
	Date time.Time
}

// take the bucket name
// look for bucket.1
// if bucket.1 (empty it)
// copy bucket to bucket.1 now
// empty bucket

// ListObjDates  return the list of objects from the
// bucket sorted by date with newest to the end
func ListObjDates(mc *minio.Client) []BucketObjects {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var bo []BucketObjects

	objectCh := mc.ListObjects(ctx, "gleaner.oihqueue", minio.ListObjectsOptions{
		Prefix:    "summoned/maspawio",
		Recursive: true,
	})
	for object := range objectCh {
		if object.Err != nil {
			fmt.Println(object.Err)
			return bo
		}

		// grab object metadata
		objInfo, err := mc.StatObject(context.Background(), "gleaner.oihqueue", object.Key, minio.StatObjectOptions{})
		if err != nil {
			fmt.Println(err)
			return bo
		}

		o := BucketObjects{Name: object.Key, Date: objInfo.LastModified}
		//fmt.Println(object.Key)
		bo = append(bo, o)
	}

	sort.Slice(bo, func(i, j int) bool { return bo[i].Date.Before(bo[j].Date) })

	return bo
}

func Copy(mc *minio.Client, b, s, d string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	objectCh := mc.ListObjects(ctx, b, minio.ListObjectsOptions{
		Prefix:    s,
		Recursive: true,
	})
	for object := range objectCh {
		if object.Err != nil {
			fmt.Println(object.Err)
		}

		// get the object name
		n := object.Key[strings.LastIndex(object.Key, "/")+1:]

		//fmt.Printf("%s/%s\n", s, n)

		srcOpts := minio.CopySrcOptions{
			Bucket: b,
			Object: fmt.Sprintf("%s/%s", s, n),
		}

		// Destination object``
		dstOpts := minio.CopyDestOptions{
			Bucket: b,
			Object: fmt.Sprintf("%s/%s", d, n),
		}

		// Copy object call
		_, err := mc.CopyObject(context.Background(), dstOpts, srcOpts)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
