package organizations

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"

	"github.com/gleanerio/gleaner/internal/common"
	"github.com/gleanerio/gleaner/internal/objects"
	"github.com/knakk/rdf"
	"github.com/xitongsys/parquet-go-source/mem"
	"github.com/xitongsys/parquet-go/writer"

	"github.com/minio/minio-go/v7"
	"github.com/spf13/viper"
)

// TEST_BuildGraph makes a graph from the Gleaner config file source
// load this to a /sources bucket (change this to sources naming convention?)
func TEST_BuildGraphMem(mc *minio.Client, v1 *viper.Viper) error {
	// var (
	// 	buf    bytes.Buffer
	// 	logger = log.New(&buf, "logger: ", log.Lshortfile)
	// )

	// read config file
	miniocfg := v1.GetStringMapString("minio")
	bucketName := miniocfg["bucket"] //   get the top level bucket for all of gleaner operations from config file

	log.Print("Building organization graph from config file")

	var domains []objects.Sources
	err := v1.UnmarshalKey("sources", &domains)
	if err != nil {
		log.Println(err)
		return err
	}

	if len(domains) == 0 {
		return nil
	}

	proc, options := common.JLDProc(v1)

	for k := range domains {
		// create new S3 file writer
		fw, err := mem.NewMemFileWriter("org.parquet", func(name string, r io.Reader) error {
			dat, err := ioutil.ReadAll(r)
			if err != nil {
				log.Printf("error reading data: %v", err)
				return err
			}

			br := bytes.NewReader(dat)

			// load to minio
			objectName := fmt.Sprintf("orgs/%s.parquet", domains[k].Name) // k is the name of the provider from config
			// contentType := "application/ld+json"

			// Upload the file with FPutObject
			_, err = mc.PutObject(context.Background(), bucketName, objectName, br, int64(br.Len()), minio.PutObjectOptions{})
			if err != nil {
				log.Printf("%s", objectName)
				log.Fatalln(err) // Fatal?   seriously?  I guess this is the object write, so the run is likely a bust at this point, but this seems a bit much still.
			}

			return err
		})
		if err != nil {
			log.Println("Can't create s3 file writer", err)
			return err
		}

		pw, err := writer.NewParquetWriter(fw, new(Qset), 4)
		if err != nil {
			log.Println("Can't create parquet writer", err)
			return err
		}

		pw.RowGroupSize = 128 * 1024 * 1024 //128M
		pw.PageSize = 8 * 1024              //8K
		// pw.CompressionType = parquet.CompressionCodec_SNAPPY

		// Sources: Name, Logo, URL, Headless, Pid

		jld, err := orggraph(domains[k])
		if err != nil {
			log.Println(err)
			return err
		}

		r, err := common.JLD2nq(jld, proc, options)
		if err != nil {
			log.Println(err)
			return err
		}

		// read rdf string line by line and feed into quad decoder

		scanner := bufio.NewScanner(strings.NewReader(r))
		for scanner.Scan() {
			rdfb := bytes.NewBufferString(scanner.Text())
			dec := rdf.NewQuadDecoder(rdfb, rdf.NQuads)

			spog, err := dec.Decode()
			if err != nil {
				log.Println(err)
				return err
			}

			qs := Qset{Subject: spog.Subj.String(), Predicate: spog.Pred.String(), Object: spog.Obj.String(), Graph: spog.Ctx.String()}

			// log.Println(qs)

			if err = pw.Write(qs); err != nil {
				log.Println("Write error", err)
				return err
			}

		}
		if err := scanner.Err(); err != nil {
			log.Println(err)
			return err
		}

		pw.Flush(true)

		if err = pw.WriteStop(); err != nil {
			log.Println("WriteStop error", err)
			return err
		}

		err = fw.Close()
		if err != nil {
			log.Println(err)
			log.Println("Error closing S3 file writer")
			return err
		}

		// delete, is this needed since we close above and have a closure call?
		if err := mem.GetMemFileFs().Remove("org.parquet"); err != nil {
			log.Printf("error removing file from memfs: %v", err)

		}
	}

	return err
}
