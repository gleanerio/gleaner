package organizations

import (
	"bytes"
	"context"
	"fmt"
	configTypes "github.com/gleanerio/gleaner/internal/config"
	log "github.com/sirupsen/logrus"
	"text/template"

	"github.com/gleanerio/gleaner/internal/common"
	"github.com/gleanerio/gleaner/internal/objects"

	"github.com/minio/minio-go/v7"
	"github.com/spf13/viper"
)

const t = `{
		"@context": {
			"@vocab": "https://schema.org/"
		},
		"@id": "https://gleaner.io/id/org/{{.Name}}",
		"@type": "Organization",
		"url": "{{.URL}}",
		"name": "{{.Name}}",
		 "identifier": {
			"@type": "PropertyValue",
			"@id": "{{.PID}}",
			"propertyID": "https://registry.identifiers.org/registry/doi",
			"url": "{{.PID}}",
			"description": "Persistent identifier for this organization"
		}
	}`

type Qset struct {
	Subject   string `parquet:"name=Subject,  type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Predicate string `parquet:"name=Predicate,  type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Object    string `parquet:"name=Object,  type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Graph     string `parquet:"name=Graph,  type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
}

// BuildGraph makes a graph from the Gleaner config file source
// load this to a /sources bucket (change this to sources naming convention?)
func BuildGraph(mc *minio.Client, v1 *viper.Viper) error {
	//var (
	//	buf    bytes.Buffer
	//	logger = log.New(&buf, "logger: ", log.Lshortfile)
	//)

	// read config file
	//miniocfg := v1.GetStringMapString("minio")
	//bucketName := miniocfg["bucket"] //   get the top level bucket for all of gleaner operations from config file
	bucketName, _ := configTypes.GetBucketName(v1)

	log.Print("Building organization graph.")
	domains := objects.SourcesAndGraphs(v1)
	proc, options := common.JLDProc(v1)

	// Sources: Name, Logo, URL, Headless, Pid
	for k := range domains {

		// log.Println(domains[k])

		jld, err := orggraph(domains[k])
		if err != nil {
			log.Println(err)
			return err
		}

		rdf, err := common.JLD2nq(jld, proc, options)
		if err != nil {
			log.Println(err)
			return err
		}

		rdfb := bytes.NewBufferString(rdf)

		// load to minio
		// orgsha := common.GetSHA(jld)
		// objectName := fmt.Sprintf("orgs/%s/%s.nq", domains[k].Name, orgsha) // k is the name of the provider from config
		objectName := fmt.Sprintf("orgs/%s.nq", domains[k].Name) // k is the name of the provider from config
		contentType := "application/ld+json"

		// Upload the file with FPutObject
		_, err = mc.PutObject(context.Background(), bucketName, objectName, rdfb, int64(rdfb.Len()), minio.PutObjectOptions{ContentType: contentType})
		if err != nil {
			logger.Printf("%s", objectName)
			return err
		}

	}

	return nil
}

func orggraph(k objects.Sources) (string, error) {
	var doc bytes.Buffer

	t, err := template.New("prov").Parse(t)
	if err != nil {
		log.Println(err)
	}

	err = t.Execute(&doc, k)
	if err != nil {
		log.Println(err)
	}

	return doc.String(), err
}

// TODO Parquet test?
// This is a test, likely should remove as this is likely a post Gleaner activity for
// something like Nabu or some other tool.

// BuildGraphPQ makes a graph from the Gleaner config file source in Parquet format
// load this to a /sources bucket (change this to sources naming convention?)
// func TEST_BuildGraphPQ(mc *minio.Client, v1 *viper.Viper) {
// 	var (
// 		buf    bytes.Buffer
// 		logger = log.New(&buf, "logger: ", log.Lshortfile)
// 	)

// 	// read config file
// 	miniocfg := v1.GetStringMapString("minio")
// 	bucketName := miniocfg["bucket"] //   get the top level bucket for all of gleaner operations from config file

// 	log.Print("Building organization graph (parquet)")
// 	domains := objects.SourcesAndGraphs(v1)
// 	proc, options := common.JLDProc(v1)

// 	for k := range domains {
// 		// get S3 info from config
// 		mcfg := v1.Sub("minio")
// 		// endpoint := fmt.Sprintf("%s:%s", mcfg.GetString("address"), mcfg.GetString("port"))
// 		accessKeyID := mcfg.GetString("accesskey")
// 		secretAccessKey := mcfg.GetString("secretkey")
// 		// useSSL := mcfg.GetBool("ssl")

// 		// Make a parquet file
// 		ctx := context.Background()
// 		region := "us-east-1"                                  // out["region"]
// 		key := fmt.Sprintf("orgs/%s.parquet", domains[k].Name) //out["object"]

// 		log.Printf("Write to %s as %s ", bucketName, key)

// 		jld, err := orggraph(domains[k])
// 		if err != nil {
// 			log.Println(err)
// 		}

// 		r, err := common.JLD2nq(jld, proc, options)
// 		if err != nil {
// 			log.Println(err)
// 		}

// 		// create new S3 file writer
// 		// TODO  WTF..  is this hard coded URL doing here?
// 		fw, err := s3.NewS3FileWriter(ctx, bucketName, key, nil, &aws.Config{Region: aws.String(region),
// 			Endpoint:    aws.String("https://192.168.86.45:32773/"),
// 			Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, "")})
// 		if err != nil {
// 			log.Println("Can't create s3 file writer", err)
// 			return
// 		}

// 		// set up parquet file
// 		pw, err := writer.NewParquetWriter(fw, new(Qset), 4)
// 		if err != nil {
// 			log.Println("Can't create parquet writer", err)
// 			return
// 		}

// 		pw.RowGroupSize = 128 * 1024 * 1024 //128M
// 		pw.PageSize = 8 * 1024              //8K
// 		// pw.CompressionType = parquet.CompressionCodec_SNAPPY

// 		// read rdf line by line and feed into quad decoder

// 		scanner := bufio.NewScanner(strings.NewReader(r))
// 		for scanner.Scan() {
// 			rdfb := bytes.NewBufferString(scanner.Text()) // WTF  why did I have r vs scanner.Text() here
// 			dec := rdf.NewQuadDecoder(rdfb, rdf.NQuads)

// 			spog, err := dec.Decode()
// 			if err != nil {
// 				logger.Println(err)
// 			}

// 			qs := Qset{Subject: spog.Subj.String(), Predicate: spog.Pred.String(), Object: spog.Obj.String(), Graph: spog.Ctx.String()}

// 			if err = pw.Write(qs); err != nil {
// 				log.Println("Write error", err)
// 			}

// 		}
// 		if err := scanner.Err(); err != nil {
// 			log.Println(err)
// 		}

// 		pw.Flush(true)

// 		if err = pw.WriteStop(); err != nil {
// 			log.Println("WriteStop error", err)
// 			return
// 		}

// 		err = fw.Close()
// 		if err != nil {
// 			log.Println(err)
// 			log.Println("Error closing S3 file writer")
// 			return
// 		}

// 	}

// }
