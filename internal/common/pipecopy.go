package common

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"sync"

	minio "github.com/minio/minio-go/v7"
)

func PipeCopyNG(name, bucket, prefix string, mc *minio.Client) error {
	log.Println("Start pipe reader / writer sequence")

	pr, pw := io.Pipe()     // TeeReader of use?
	lwg := sync.WaitGroup{} // work group for the pipe writes...
	lwg.Add(2)

	// params for list objects calls
	doneCh := make(chan struct{}) // , N) Create a done channel to control 'ListObjectsV2' go routine.
	defer close(doneCh)           // Indicate to our routine to exit cleanly upon return.
	isRecursive := true

	go func() {
		defer lwg.Done()
		defer pw.Close()
		objectCh := mc.ListObjects(context.Background(), bucket, minio.ListObjectsOptions{Prefix: prefix, Recursive: isRecursive})

		// for object := range mc.ListObjects(context.Background(), bucket, minio.ListObjectsOptions{Prefix: prefix, Recursive: isRecursive}, doneCh) {
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

			pw.Write(b.Bytes())
		}

	}()

	// log.Printf("%s_graph.nq", name)

	// go function to write to minio from pipe
	go func() {
		defer lwg.Done()
		_, err := mc.PutObject(context.Background(), "gleaner", name, pr, -1, minio.PutObjectOptions{})
		if err != nil {
			log.Println(err)
		}
	}()

	// Note: We can also make a file and pipe write to that, keep this code around in case
	// f, err := os.Create(fmt.Sprintf("%s_graph.nq", prefix))  // needs a f.Close() later
	// if err != nil {
	// 	log.Println(err)
	// }
	// go function to write to file from pipe
	// go func() {
	// 	defer lwg.Done()
	// 	if _, err := io.Copy(f, pr); err != nil {
	// 		log.Fatal(err)
	// 	}
	// }()

	lwg.Wait() // wait for the pipe read writes to finish
	pw.Close()
	pr.Close()

	return nil
}
