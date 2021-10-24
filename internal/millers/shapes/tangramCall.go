package shapes

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"strings"

	"context"
	"github.com/earthcubearchitecture-project418/gleaner/internal/common"
	"github.com/earthcubearchitecture-project418/gleaner/internal/millers/graph"
	minio "github.com/minio/minio-go/v7"
	"github.com/piprate/json-gold/ld"
	"github.com/spf13/viper"
)

// Call the SHACL service container (or cloud instance) // TODO: service URL needs to be in the config file!
func shaclTestNG(v1 *viper.Viper, bucketname, prefix string, mc *minio.Client, object, shape minio.ObjectInfo, proc *ld.JsonLdProcessor, options *ld.JsonLdOptions) (string, error) {
	key := object.Key // replace if new function idea works..

	// Read the object bytes (our data graoh)
	//fo, err := mc.GetObject(bucketname, object.Key, minio.GetObjectOptions{})
	fo, err := mc.GetObject(context.Background(), bucketname, object.Key, minio.GetObjectOptions{})
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	var b bytes.Buffer
	bw := bufio.NewWriter(&b)

	_, err = io.Copy(bw, fo)
	if err != nil {
		log.Println(err)
	}

	// TODO  this is a waste to read the same bytes N times!   read early and pass a pointer!
	// Read the object bytes (our data graoh)
	//so, err := mc.GetObject("gleaner", shape.Key, minio.GetObjectOptions{})
	so, err := mc.GetObject(context.Background(), "gleaner", shape.Key, minio.GetObjectOptions{})
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	var sb bytes.Buffer
	sbw := bufio.NewWriter(&sb)

	_, err = io.Copy(sbw, so)
	if err != nil {
		log.Println(err)
	}

	// TODO
	// Process the bytes in b to RDF (with randomized blank nodes)
	// rdf, err := jld2nq(string(b.Bytes()), key, proc, options)
	// if err != nil {
	// 	return key, err
	// }

	// rdf := "The results of the SHACL call in nquads"
	// rdfubn := GlobalUniqueBNodes(rdf)
	// rdfubn := "blank node fixed RDF if I can't skolemize in Tangram"

	//log.Println(string(b.Bytes()))
	//log.Println("------------------")
	//log.Println(string(sb.Bytes()))

	// get the URL from viper object
	mcfg := v1.GetStringMapString("shaclservice")

	// TODO
	// resolve how call
	rdfubn, err := shaclCallNG(mcfg["url"], string(b.Bytes()), string(sb.Bytes()))
	if err != nil {
		log.Print(err)
	}

	// TODO we have our nt from SHACL, but it needs some extra info to let us
	// build reports.  The response is ntriples, so easy to find the subject
	// IRI.   On our end we need the @id or schema:url.  I hate doing another
	// heavy frame. Can I get the value earlier in the chain?

	milledkey := strings.ReplaceAll(key, ".jsonld", ".rdf")
	milledkey = strings.ReplaceAll(milledkey, "summoned/", "")

	// make an object with prefix like  scienceorg-dg/objectname.rdf  (where is had .jsonld before)
	// objectName := fmt.Sprintf("%s-shacl/%s", prefix, strings.ReplaceAll(key, ".jsonld", ".rdf"))

	objectName := fmt.Sprintf("%s/%s", prefix, milledkey)

	//contentType := "application/ld+json"
	usermeta := make(map[string]string) // what do I want to know?
	usermeta["origfile"] = key
	//		usermeta["url"] = urlloc
	//		usermeta["sha1"] = sha
	//		bucketName := "gleaner-summoned" //   fmt.Sprintf("gleaner-summoned/%s", k) // old was just k

	// Upload the file
	_, err = graph.LoadToMinio(rdfubn, "gleaner", objectName, mc)
	if err != nil {
		return objectName, err
	}

	return objectName, nil
}

// Call the SHACL service container (or cloud instance) // TODO: service URL needs to be in the config file!
func shaclCallNG(url, dg, sg string) (string, error) {
	// datagraph, err := millerutils.JSONLDToTTL(dg, urlval)
	// if err != nil {
	// 	log.Printf("Error in the jsonld write... %v\n", err)
	// 	log.Printf("Nothing to do..   going home")
	// 	return 0
	// }

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	// writer.WriteField("datagraph", urlval)
	// writer.WriteField("shapegraph", sgkey)
	writer.WriteField("fmt", "nt")

	//part, err := writer.CreateFormFile("datagraph", "datagraph")
	part, err := writer.CreateFormFile("dg", "datagraph")
	if err != nil {
		return "", err
	}
	_, err = io.Copy(part, strings.NewReader(dg))
	if err != nil {
		return "", err
	}

	//part, err = writer.CreateFormFile("shapegraph", "shapegraph")
	part, err = writer.CreateFormFile("sg", "shapegraph")
	if err != nil {
		return "", err
	}
	_, err = io.Copy(part, strings.NewReader(sg))
	if err != nil {
		return "", err
	}

	err = writer.Close()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", "EarthCube_DataBot/1.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(b), err //  we will return the bytes count we write...
}

// DEPRECATED CODE BELOW..   will be replaced
// Call the SHACL service container (or cloud instance) // TODO: service URL needs to be in the config file!
func shaclTest(urlval, dg, sgkey, sg string, gb *common.Buffer) int {
	// datagraph, err := millerutils.JSONLDToTTL(dg, urlval)
	// if err != nil {
	// 	log.Printf("Error in the jsonld write... %v\n", err)
	// 	log.Printf("Nothing to do..   going home")
	// 	return 0
	// }

	url := "http://localhost:8080/uploader" // TODO this should be set in the config file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("datagraph", urlval)
	writer.WriteField("shapegraph", sgkey)

	part, err := writer.CreateFormFile("datagraph", "datagraph")
	if err != nil {
		log.Println(err)
	}
	_, err = io.Copy(part, strings.NewReader(dg))
	if err != nil {
		log.Println(err)
	}

	part, err = writer.CreateFormFile("shapegraph", "shapegraph")
	if err != nil {
		log.Println(err)
	}
	_, err = io.Copy(part, strings.NewReader(sg))
	if err != nil {
		log.Println(err)
	}

	err = writer.Close()
	if err != nil {
		log.Println(err)
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		log.Println(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", "EarthCube_DataBot/1.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}

	// write result to buffer
	len, err := gb.Write(b)
	if err != nil {
		log.Printf("error in the buffer write... %v\n", err)
	}

	return len //  we will return the bytes count we write...
}
