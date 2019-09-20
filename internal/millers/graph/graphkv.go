package graph

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"earthcube.org/Project418/gleaner/internal/common"
	"earthcube.org/Project418/gleaner/internal/millers/millerutils"
	"earthcube.org/Project418/gleaner/pkg/utils"

	"github.com/knakk/rdf"
	minio "github.com/minio/minio-go"
	bolt "go.etcd.io/bbolt"
)

// MillerSetup issues a go call for each domain in the ocnfig file to mill the graph
func MillerSetup(mc *minio.Client, b []string, cs utils.Config) {
	//	uiprogress.Start()
	wg := sync.WaitGroup{}

	for k := range b {
		log.Printf("Queuing URLs for %s \n", b[k])
		// 	go getDomain(mc, m, k, &wg)
		go Miller(mc, b[k], cs, &wg) // kv based function (disk based with memory mapping)
	}

	time.Sleep(2 * time.Second)
	wg.Wait()
	//	uiprogress.Stop()
}

// Miller (dev version to deal with memory and scale isues)
func Miller(mc *minio.Client, prefix string, cs utils.Config, wg *sync.WaitGroup) {
	wg.Add(1)

	doneCh := make(chan struct{}) // , N) Create a done channel to control 'ListObjectsV2' go routine.
	defer close(doneCh)           // Indicate to our routine to exit cleanly upon return.
	isRecursive := true
	bucketname := "gleaner-summoned"
	objectCh := mc.ListObjectsV2(bucketname, prefix, isRecursive, doneCh)

	db := kvclient(prefix)
	defer db.Close()

	log.Println(prefix)

	for object := range objectCh {
		if object.Err != nil {
			log.Println(object.Err)
		}

		fo, err := mc.GetObject(bucketname, object.Key, minio.GetObjectOptions{})
		if err != nil {
			log.Println(err)
		}
		oi, err := fo.Stat()
		if err != nil {
			log.Println("Issue with reading an object..  should I just fail on this to make sure?")
		}

		var urlval, sha1val string
		if len(oi.Metadata["X-Amz-Meta-Url"]) > 0 {
			urlval = oi.Metadata["X-Amz-Meta-Url"][0] // also have  X-Amz-Meta-Sha1
		}
		if len(oi.Metadata["X-Amz-Meta-Sha1"]) > 0 {
			sha1val = oi.Metadata["X-Amz-Meta-Sha1"][0]
		}

		buf := new(bytes.Buffer)
		buf.ReadFrom(fo)
		jld := buf.String() // Does a complete copy of the bytes in the buffer.

		cb := new(common.Buffer) // TODO..   really just a bytes buffer should be used

		_ = millerutils.Jsl2graph(bucketname, object.Key, urlval, sha1val, jld, cb)

		good, bad, err := graphSplit(cb, bucketname)

		db.Update(func(tx *bolt.Tx) error {
			b2 := tx.Bucket([]byte("GoodTriples"))
			if err != nil {
				log.Println(err)
			}
			err = b2.Put([]byte(urlval), []byte(good))
			return err
		})

		db.Update(func(tx *bolt.Tx) error {
			b3 := tx.Bucket([]byte("BadTriples"))
			if err != nil {
				log.Println(err)
			}
			err = b3.Put([]byte(urlval), []byte(bad))
			return err
		})

		cb.Reset()
	}

	log.Println("Start pipe reader / writer sequence")
	// ref io.Pipe https://stackoverflow.com/questions/37645869/how-to-deal-with-io-eof-in-a-bytes-buffer-stream
	// https://zupzup.org/io-pipe-go/
	// https://rodaine.com/2015/04/async-split-io-reader-in-golang/
	pr, pw := io.Pipe() // TeeReader of use?

	// work group for the pipe writes...
	lwg := sync.WaitGroup{}
	lwg.Add(2)

	go func() {
		defer lwg.Done()
		defer pw.Close()
		db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("GoodTriples"))
			c := b.Cursor()
			for k, v := c.First(); k != nil; k, v = c.Next() {
				pw.Write(v)
			}
			return nil
		})
	}()

	// file object needed by pipe to file go function
	// f, err := os.Create(fmt.Sprintf("%s_graph.nq", prefix))
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

	// go function to write to minio from pipe
	go func() {
		defer lwg.Done()
		_, err := mc.PutObject("gleaner-milled", prefix, pr, -1, minio.PutObjectOptions{})
		if err != nil {
			log.Println(err)
		}
	}()

	lwg.Wait() // wait for the pipe read writes to finish
	pw.Close()
	pr.Close()
	// f.Close() // close file object needed by pipe to file go function

	wg.Done()
}

func graphSplit(gb *common.Buffer, bucketname string) (string, string, error) {
	var err error
	scanner := bufio.NewScanner(gb) // rdf is already a pointer
	good := bytes.NewBuffer(make([]byte, 0))
	bad := bytes.NewBuffer(make([]byte, 0))
	for scanner.Scan() {
		if len(scanner.Text()) > 2 {
			nq, e := goodTriples(scanner.Text(), fmt.Sprintf("http://earthcube.org/%s", bucketname))
			if e == nil {
				_, err = good.Write([]byte(nq))
			}
			if e != nil {
				_, err = bad.Write([]byte(fmt.Sprintf("%s :Error: %s\n", strings.TrimSuffix(scanner.Text(), "\n"), e)))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}

	return good.String(), bad.String(), err
}

// TODO  convert this to use a bytes.Buffer  (or better a pointer to that)
func goodTriples(f, c string) (string, error) {
	// fmt.Printf("Trying: %s \n", f)
	dec := rdf.NewTripleDecoder(strings.NewReader(f), rdf.NTriples)
	triple, err := dec.Decode()
	if err != nil {
		log.Printf("Error decoding triples: %v\n", err)
		return "", err
	}

	// enc := rdf.NewQuadEncoder(outFile, rdf.NQuads)
	q, err := makeQuad(triple, c)
	if err != nil {
		return "", err
	}

	return q, err // q is alread a string..
}

// makeQuad make a quad from a triple and a context string
func makeQuad(t rdf.Triple, c string) (string, error) {
	newctx, err := rdf.NewIRI(c)
	if err != nil {
		return "", err
	}
	ctx := rdf.Context(newctx)
	q := rdf.Quad{Triple: t, Ctx: ctx}
	qs := q.Serialize(rdf.NQuads)

	return qs, err
}

// pass in a bucket name here to make several of these
// and use trhe go func pattern from the summoner
// to do these graph builds in parrallel
// Could 1 db but might have write collisions more then
func kvclient(name string) *bolt.DB {

	dir, err := ioutil.TempDir("", name) // emptry string puts tmp dir in os.TempDir
	if err != nil {
		log.Fatal(err)

	}
	defer os.RemoveAll(dir)

	db, err := bolt.Open(fmt.Sprintf("%s/%s.db", dir, name), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("JSONLD"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		_, err = tx.CreateBucket([]byte("GoodTriples"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		_, err = tx.CreateBucket([]byte("BadTriples"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	return db

}

// https://play.golang.org/p/c0fLEI350w
// func background(mc *minio.Client, r io.Reader, o string) {
// 	buf := make([]byte, 64)

// 	for {
// 		_, err := r.Read(buf)
// 		if err != nil {
// 			fmt.Println(err.Error())
// 			return
// 		}

// 		ior := bytes.NewReader(buf)

// 		_, err = mc.PutObject("test", o, ior, -1, minio.PutObjectOptions{})
// 		if err != nil {
// 			log.Fatalln(err)
// 		}
// 	}

// }
