package objects

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/gleanerio/gleaner/nabu/pkg/nabuinternal/graph"
	log "github.com/sirupsen/logrus"

	"github.com/minio/minio-go/v7"
)

func getLastElement(s string) string {
	parts := strings.Split(s, "/")
	return parts[len(parts)-1]
}

// GenerateDateHash generates a unique hash based on the current date and time.
func generateDateHash() string {
	// Get the current date and time
	now := time.Now()

	// Format the date and time as a string
	dateString := now.Format("2006-01-02 15:04:05")

	// Create a SHA256 hash
	hash := sha256.New()
	hash.Write([]byte(dateString))

	// Convert the hash to a hex string
	hashString := hex.EncodeToString(hash.Sum(nil))

	return hashString
}

// PipeCopy writes a new object based on an prefix, this function assumes the objects are valid when concatenated
// v1:  viper config object
// mc:  minio client pointer
// name:  name of the NEW object
// bucket:  source bucket  (and target bucket)
// prefix:  source prefix
// destprefix:   destination prefix
// sf: boolean to declare if single file or not.   If so, skip skolimization since JSON-LD library output is enough
func PipeCopy(v1 *viper.Viper, mc *minio.Client, name, bucket, prefix, destprefix string) error {
	orgname := v1.GetString("implementation_network.orgname")
	log.Printf("PipeCopy with name: %s   bucket: %s  prefix: %s  org name: %s", name, bucket, prefix, orgname)

	pr, pw := io.Pipe()     // TeeReader of use?
	lwg := sync.WaitGroup{} // work group for the pipe writes...
	lwg.Add(2)

	// params for list objects calls
	doneCh := make(chan struct{}) // , N) Create a done channel to control 'ListObjectsV2' go routine.
	defer close(doneCh)           // Indicate to our routine to exit cleanly upon return.
	isRecursive := true

	//log.Printf("Bulkfile name: %s_graph.nq", name)

	go func() {
		defer lwg.Done()
		defer func(pw *io.PipeWriter) {
			err := pw.Close()
			if err != nil {
			}
		}(pw)

		clen := 0
		sf := false
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		lenCh := mc.ListObjects(ctx, bucket, minio.ListObjectsOptions{Prefix: prefix, Recursive: isRecursive})
		for _ = range lenCh {
			clen = clen + 1
		}
		if clen == 1 {
			sf = true
		}
		log.Printf("\nChannel/object length: %d\n", clen)
		log.Printf("Single file mode set: %t", sf)

		objectCh := mc.ListObjects(context.Background(), bucket, minio.ListObjectsOptions{Prefix: prefix, Recursive: isRecursive})

		lastProcessed := false
		idList := make([]string, 0)
		for object := range objectCh {
			fo, err := mc.GetObject(context.Background(), bucket, object.Key, minio.GetObjectOptions{})
			if err != nil {
				log.Errorf(" failed to read object %s %s ", object.Key, err)
				continue
			}
			log.Tracef(" processing  object %s  ", object.Key)
			var b bytes.Buffer
			bw := bufio.NewWriter(&b)
			_, err = io.Copy(bw, fo)
			if err != nil {
				log.Errorf(" failed to read object %s %s ", object.Key, err)
				log.Println(err)
				continue
			}
			s := string(b.Bytes())
			nq := ""
			if strings.HasSuffix(object.Key, ".nq") {
				nq = s
			} else {
				nq, err = graph.JSONLDToNQ(v1, s)
				if err != nil {
					log.Errorf(" failed to convert to NQ %s %s ", object.Key, err)
					continue
				}
			}
			var snq string
			if sf {
				snq = nq
			} else {
				snq, err = graph.Skolemization(nq, object.Key)
				if err != nil {
					log.Errorf(" failed Skolemization %s %s ", object.Key, err)
					continue
				}
			}
			ctx, err := graph.MakeURN(v1, object.Key)
			if err != nil {
				log.Errorf(" failed MakeURN %s %s ", object.Key, err)
				continue
			}
			csnq, err := graph.NtToNq(snq, ctx)
			if err != nil {
				log.Errorf(" failed NtToNq %s %s ", object.Key, err)
				continue
			}
			_, err = pw.Write([]byte(csnq))
			if err != nil {
				log.Errorf(" failed pipe write %s %s ", object.Key, err)
				continue
			}
			idList = append(idList, ctx)
			lastProcessed = true
		}

		// Once we are done with the loop, put in the triples to associate all the graphURIs with the org.
		if lastProcessed {

			data := `<urn:gleaner.io:` + orgname + `:datacatalog> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.org/DataCatalog> .
<urn:gleaner.io:` + orgname + `:datacatalog> <https://schema.org/description> "GleanerIO Nabu generated catalog" .
<urn:gleaner.io:` + orgname + `:datacatalog>  <https://schema.org/dateCreated> "` + time.Now().Format("2006-01-02 15:04:05") + `" .
<urn:gleaner.io:` + orgname + `:datacatalog> <https://schema.org/provider> <urn:gleaner.io:` + orgname + `:provider> .
<urn:gleaner.io:` + orgname + `:datacatalog> <https://schema.org/publisher> <urn:gleaner.io:` + getLastElement(prefix) + `:publisher> .
<urn:gleaner.io:` + orgname + `:provider> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.org/Organization> .
<urn:gleaner.io:` + orgname + `:provider> <https://schema.org/name> "` + orgname + `" .
<urn:gleaner.io:` + getLastElement(prefix) + `:publisher> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.org/Organization> .
<urn:gleaner.io:` + getLastElement(prefix) + `:publisher> <https://schema.org/name> "` + getLastElement(prefix) + `" .
`

			for _, item := range idList {
				data += `<urn:gleaner.io:` + orgname + `:datacatalog> <https://schema.org/dataset> <` + item + `> .` + "\n"
			}

			namedgraph := "urn:gleaner.io:" + orgname + ":" + getLastElement(prefix) + ":datacatalog:" + generateDateHash()
			sdata, err := graph.NtToNq(data, namedgraph)

			// Perform the final write to the pipe here
			// ilstr := strings.Join(idList, ",")
			_, err = pw.Write([]byte(sdata))
			if err != nil {
				log.Println(err)
			}
		}
	}()

	// go function to write to minio from pipe
	go func() {
		defer lwg.Done()
		_, err := mc.PutObject(context.Background(), bucket, fmt.Sprintf("%s/%s", destprefix, name), pr, -1, minio.PutObjectOptions{})
		//_, err := mc.PutObject(context.Background(), bucket, fmt.Sprintf("%s/%s", prefix, name), pr, -1, minio.PutObjectOptions{})
		if err != nil {
			log.Errorf(" failed PutObject bucket: %s  %s/%s ", bucket, destprefix, name)
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
