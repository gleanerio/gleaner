package graph

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"earthcube.org/Project418/gleaner/internal/common"
	"github.com/blevesearch/bleve"
	"github.com/knakk/rdf"
	minio "github.com/minio/minio-go"
	"github.com/piprate/json-gold/ld"
	"github.com/rs/xid"
)

var RunDir string // could inlcude an init() func to check this is set and default or error

// Jsl2graph is a simple function to do stuff  :)
func Jsl2graph(bucketname, key, urlval, sha1val, jsonld string, gb *common.Buffer) int {
	nq, err := JSONLDToNQ(jsonld, urlval) // TODO replace with NQ from isValid function..  saving time..
	if err != nil {
		log.Printf("error in the jsonld write... %v\n", err)
	}

	rdf := GlobalUniqueBNodes(nq) // unique bnodes
	lpt := LPtriples(rdf, urlval) // associate landing page URL with all unique subject URIs and subject bnodes in graph

	nt := fmt.Sprint("\n" + rdf + "\n" + lpt)

	len, err := gb.Write([]byte(nt))
	if err != nil {
		log.Printf("error in  buffer write... %v\n", err)
	}

	return len //  we will return the bytes count we write...
}

// TODO look for <s> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://schema.org/Dataset> .
func LPtriples(t, urlval string) string {
	scanner := bufio.NewScanner(strings.NewReader(t))

	us := []string{}
	for scanner.Scan() {
		split := strings.Split(scanner.Text(), " ")
		us = appendIfMissing(us, split[0])
	}

	nt := []string{}
	for i := range us {
		t := fmt.Sprintf("%s <http://www.w3.org/2000/01/rdf-schema#seeAlso> <%s> .", us[i], urlval)
		nt = append(nt, t)
	}

	lpt := strings.Join(nt, "\n")

	return lpt
}

func appendIfMissing(slice []string, i string) []string {
	for _, ele := range slice {
		if ele == i {
			return slice
		}
	}
	return append(slice, i)
}

// JSONLDToNQ converts JSON-LD documents to NQuads
func JSONLDToNQ(jsonld, urlval string) (string, error) {
	proc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")
	options.Format = "application/nquads"

	var myInterface interface{}
	err := json.Unmarshal([]byte(jsonld), &myInterface)
	if err != nil {
		log.Printf("Error when transforming %s JSON-LD document to interface: %v", urlval, err)
		return "", err
	}

	nq, err := proc.ToRDF(myInterface, options)
	if err != nil {
		log.Printf("Error when transforming %s  JSON-LD document to RDF: %v", urlval, err)
		return "", err
	}

	return nq.(string), err
}

// JSONLDToTTL Relates the SHACL support..  (REMOVE THIS...   put JSON-LD -> turtle in the java service)
func JSONLDToTTL(jsonld, urlval string) (string, error) {
	// Sad that rdf2go has a bug in jsonld around blank nodes...
	// I can convert to NQ above..   I guess use knakk to then convert to turtle  (since json-ld gold
	// also does not support converting to TTL.

	nq, err := JSONLDToNQ(jsonld, urlval)
	if err != nil {
		log.Println(err)
		return "", err
	}

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	dec := rdf.NewTripleDecoder(strings.NewReader(nq), rdf.NTriples)
	tr, err := dec.DecodeAll()
	if err != nil {
		log.Println(err)
	}

	enc := rdf.NewTripleEncoder(writer, rdf.Turtle)
	err = enc.EncodeAll(tr)
	if err != nil {
		log.Println(err)
	}
	writer.Flush()

	return b.String(), err
}

// GlobalUniqueBNodes should NOT be here.  However at this time the state of RDF stores in golang
// doesn't include one that can deal with bnodes.  So, I have to ensure they are GUIDs going in or
// the all get named _:b# where 3 always indexes from 0   (I pray I can remove this someday soon!)
func GlobalUniqueBNodes(nq string) string {
	scanner := bufio.NewScanner(strings.NewReader(nq))
	// make a map here to hold our old to new map
	m := make(map[string]string)

	for scanner.Scan() {
		//fmt.Println(scanner.Text())
		// parse the line
		split := strings.Split(scanner.Text(), " ")
		sold := split[0]
		oold := split[2]

		if strings.HasPrefix(sold, "_:") { // we are a blank node
			// check map to see if we have this in our value already
			if _, ok := m[sold]; ok {
				// fmt.Printf("We had %s, already\n", sold)
			} else {
				guid := xid.New()
				snew := fmt.Sprintf("_:b%s", guid.String())
				m[sold] = snew
			}
		}

		// scan the object nodes too.. though we should find nothing here.. the above wouldn't
		// eventually find
		if strings.HasPrefix(oold, "_:") { // we are a blank node
			// check map to see if we have this in our value already
			if _, ok := m[oold]; ok {
				// fmt.Printf("We had %s, already\n", oold)
			} else {
				guid := xid.New()
				onew := fmt.Sprintf("_:b%s", guid.String())
				m[oold] = onew
			}
		}
		// triple := tripleBuilder(split[0], split[1], split[3])
		// fmt.Println(triple)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	//fmt.Println(m)

	filebytes := []byte(nq)

	for k, v := range m {
		// fmt.Printf("Replace %s with %v \n", k, v)
		filebytes = bytes.Replace(filebytes, []byte(k), []byte(v), -1)
	}

	return string(filebytes)
}

// NewinitBleve Initialize the text index  // this function needs some attention (of course they all do)
func NewinitBleve(filename string) string {

	path := fmt.Sprintf("%s/bleve", RunDir)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, os.ModePerm)
	}

	filepath := fmt.Sprintf("%s/bleve/%s", RunDir, filename)
	mapping := bleve.NewIndexMapping()
	index, berr := bleve.New(filepath, mapping)
	if berr != nil {
		log.Printf("Bleve error making index %v \n", berr)
	}
	index.Close()

	return filepath
}

// LoadToMinio loads jsonld into the specified bucket
func LoadToMinio(jsonld, bucketName, objectName string, mc *minio.Client) (int64, error) {

	// set up some elements for PutObject
	contentType := "application/ld+json"
	b := bytes.NewBufferString(jsonld)
	usermeta := make(map[string]string) // what do I want to know?
	// usermeta["url"] = urlloc
	// usermeta["sha1"] = bss

	log.Println(bucketName)
	// Upload the zip file with FPutObject
	n, err := mc.PutObject(bucketName, objectName, b, int64(b.Len()), minio.PutObjectOptions{ContentType: contentType, UserMetadata: usermeta})
	if err != nil {
		log.Printf("%s", objectName)
		log.Fatalln(err)
	}

	// log.Printf("#%d Uploaded Bucket:%s File:%s Size %d\n", i, bucketName, objectName, n)

	return n, nil
}

// WriteRDF save the RDF graph to a file
func WriteRDF(rdf string, prefix string) (int, error) {
	// for now just append to a file..   later I will send to a triple store
	// If the file doesn't exist, create it, or append to the file

	// check if our graphs directory exists..  make it if it doesn't
	path := fmt.Sprintf("%s/graphs", RunDir)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, os.ModePerm)
	}

	f, err := os.OpenFile(fmt.Sprintf("%s/graphs/%s.n3", RunDir, prefix), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	fl, err := f.Write([]byte(rdf))
	if err != nil {
		log.Fatal(err)
	}

	if err := f.Close(); err != nil {
		log.Fatal(err)
	}

	return fl, err // always nil,  we will never get here with FATAL..   leave for test..  but later remove to log only
}
