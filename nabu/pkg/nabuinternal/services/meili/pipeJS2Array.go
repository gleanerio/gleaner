package meili

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/sjson"

	"github.com/minio/minio-go/v7"
)

// TODO this function is too specialized (due the ID setting) to be in object
// it needs to be moved to meili

// ToJSONArray writes a new object based on an prefix it concat to that object
// name:  name of the NEW object
// bucket:  source bucket  (and target bucket)
// prefix:  source prefix
// mc:  minio client pointer
func ToJSONArray(name, bucket, prefix string, mc *minio.Client) error {
	log.Printf("ToJSONArray with name: %s   bucket: %s  prefix: %s", name, bucket, prefix)

	pr, pw := io.Pipe()     // TeeReader of use?
	lwg := sync.WaitGroup{} // work group for the pipe writes...
	lwg.Add(2)

	// params for list objects calls
	doneCh := make(chan struct{}) // , N) Create a done channel to control 'ListObjectsV2' go routine.
	defer close(doneCh)           // Indicate to our routine to exit cleanly upon return.
	isRecursive := true

	go func() {
		defer lwg.Done()
		defer func(pw *io.PipeWriter) {
			err := pw.Close()
			if err != nil {
			}
		}(pw)

		_, err := pw.Write([]byte("["))
		if err != nil {
			return
		}

		objectCh := mc.ListObjects(context.Background(), bucket, minio.ListObjectsOptions{Prefix: prefix, Recursive: isRecursive})

		// for object := range mc.ListObjects(context.Background(), bucket, minio.ListObjectsOptions{Prefix: prefix, Recursive: isRecursive}, doneCh) {
		first := true
		for object := range objectCh {
			fo, err := mc.GetObject(context.Background(), bucket, object.Key, minio.GetObjectOptions{})
			if err != nil {
				fmt.Println(err)
			}

			var b bytes.Buffer
			bw := bufio.NewWriter(&b)

			_, err = io.Copy(bw, fo)
			if err != nil {
				log.Println(err)
			}

			// Meili start  this could be an external "mod" function
			// Build ID entry to support meilie and since this ID is used only to associate a record back to the graph
			s := string(b.Bytes())

			fp := filepath.Base(object.Key)
			nns := strings.TrimSuffix(fp, path.Ext(fp))

			fv, err := sjson.Set(s, "@id", nns)
			if err != nil {
				log.Println(err)
			}
			fv2, err := sjson.Set(fv, "id", nns)
			if err != nil {
				log.Println(err)
			}
			// Meili end

			// Do not want a "," at the end of the last element, so write it prior to the record and skip the first record
			if first {
				first = false
			} else {
				_, err = pw.Write([]byte(","))
				if err != nil {
					return
				}
			}

			//_, err = pw.Write(b.Bytes())
			_, err = pw.Write([]byte(fv2))
			if err != nil {
				return
			}

		}

		_, err = pw.Write([]byte("]"))
		if err != nil {
			return
		}

	}()

	// go function to write to minio from pipe
	go func() {
		defer lwg.Done()
		_, err := mc.PutObject(context.Background(), bucket, fmt.Sprintf("%s/%s", prefix, name), pr, -1, minio.PutObjectOptions{})
		if err != nil {
			log.Println(err)
			return
		}
	}()

	lwg.Wait() // wait for the pipe read writes to finish
	err := pw.Close()
	if err != nil {
		return err
	}
	err = pr.Close()
	if err != nil {
		return err
	}

	return nil
}
