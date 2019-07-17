package graph

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	"earthcube.org/Project418/gleaner/internal/common"
	"earthcube.org/Project418/gleaner/internal/millers/millerutils"
	"earthcube.org/Project418/gleaner/pkg/utils"

	"github.com/knakk/rdf"
	minio "github.com/minio/minio-go"
	bolt "go.etcd.io/bbolt"
)

// Miller (dev version to deal with memory and scale isues)
func Miller(mc *minio.Client, prefix string, cs utils.Config) {
	doneCh := make(chan struct{}) // Create a done channel to control 'ListObjectsV2' go routine.
	defer close(doneCh)           // Indicate to our routine to exit cleanly upon return.
	isRecursive := true
	bucketname := "gleaner-summoned"
	objectCh := mc.ListObjectsV2(bucketname, prefix, isRecursive, doneCh)

	db := kvclient()
	defer db.Close()

	for object := range objectCh {
		if object.Err != nil {
			fmt.Println(object.Err)
		}

		fo, err := mc.GetObject(bucketname, object.Key, minio.GetObjectOptions{})
		if err != nil {
			fmt.Println(err)
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
		//rb, err := ioutil.ReadAll(cb) // read buffer to []byte
		//if err != nil {
		//		log.Println(err)
		//		}

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

	// b2 := tx.Bucket([]byte("GoodTriples"))
	// dbs := db.Stats()
	// db.View(func(tx *bolt.Tx) error {
	// 	// Assume bucket exists and has keys
	// 	b := tx.Bucket([]byte("GoodTriples"))

	// 	b.ForEach(func(k, v []byte) error {
	// 		fmt.Printf("key=%s, value=%s\n", k, v)
	// 		return nil
	// 	})
	// 	return nil
	// })

	fmt.Println("cp0")

	// ref io.Pipe https://stackoverflow.com/questions/37645869/how-to-deal-with-io-eof-in-a-bytes-buffer-stream
	// https://zupzup.org/io-pipe-go/
	// https://rodaine.com/2015/04/async-split-io-reader-in-golang/
	pr, pw := io.Pipe() // TeeReader of use?
	fmt.Println("cp1")
	// we need to wait for everything to be done

	// TODO
	//  No pooint for this to be a work group?
	// or do 2 seperate ones....

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
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

	// TODO replace os.Stdout with a file writer
	f, err := os.Create("./TESTOUT.txt")
	if err != nil {
		log.Println(err)
	}

	go func() {
		defer wg.Done()
		// read from the PipeReader to stdout
		if _, err := io.Copy(f, pr); err != nil {
			log.Fatal(err)
		}
	}()

	wg.Wait()

	/*
		db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("GoodTriples"))
			c := b.Cursor()

			fmt.Println("cp2")
			for k, v := c.First(); k != nil; k, v = c.Next() {
				//pw.Write(v)
				//_, err := tx.WriteTo(pw)
				//if err != nil {
				//		log.Println(err)
				//}
				//fmt.Fprint(pw, v)
				// fmt.Fprintf(pw, "%d)teststring\n", i)
				fmt.Printf("key=%s, value=%d\n", k, len(v))
			}

			return nil
		})
	*/

	//pw.Close()
	fmt.Println("cp3")
	// background(mc, pr)

}

// https://play.golang.org/p/c0fLEI350w
func background(mc *minio.Client, r io.Reader) {
	buf := make([]byte, 64)

	for {
		_, err := r.Read(buf)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		ior := bytes.NewReader(buf)

		_, err = mc.PutObject("test", "test-object2", ior, -1, minio.PutObjectOptions{})
		if err != nil {
			log.Fatalln(err)
		}

		// n, err := r.Read(buf)
		// if err != nil {
		// 	fmt.Print(err.Error())
		// 	//return
		// }
		// fmt.Print(string(buf[:n]))
	}
	// log.Println("Uploaded", "my-objectname", " of size: ", n, "Successfully.")
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

func kvclient() *bolt.DB {
	// Do this in main and pass the reference later?
	db, err := bolt.Open("gleaner.db", 0600, nil)
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
