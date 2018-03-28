package acquire

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"earthcube.org/Project418/gleaner/summoner/sitemaps"
	"earthcube.org/Project418/gleaner/summoner/utils"
	"github.com/kazarena/json-gold/ld"
	minio "github.com/minio/minio-go"
)

// Sources is a struct holding the metadata associated with the sites to harvest
type Sources struct {
	URL           string
	Headless      bool
	SitemapFormat string
	Active        bool
}

// LoadToMinio loads jsonld into the specified bucket
func LoadToMinio(jsonld, bucketName, urlloc string, minioClient *minio.Client, i int) (string, string, error) {
	// get sha1 of the JSONLD..  it's a nice ID
	h := sha1.New()
	h.Write([]byte(jsonld))
	bs := h.Sum(nil)
	bss := fmt.Sprintf("%x", bs) // better way to convert bs hex string to string?

	// objectName := fmt.Sprintf("%s/%s.jsonld", up.Path, bss)
	objectName := fmt.Sprintf("%s.jsonld", bss)
	contentType := "application/ld+json"
	b := bytes.NewBufferString(jsonld)

	usermeta := make(map[string]string) // what do I want to know?
	usermeta["url"] = urlloc
	usermeta["sha1"] = bss
	// bucketName := k

	// Upload the zip file with FPutObject
	n, err := minioClient.PutObject(bucketName, objectName, b, int64(b.Len()), minio.PutObjectOptions{ContentType: contentType, UserMetadata: usermeta})
	if err != nil {
		log.Printf("%s", objectName)
		log.Fatalln(err)
	}

	log.Printf("#%d Uploaded Bucket:%s File:%s Size %d\n", i, bucketName, objectName, n)

	return urlloc, objectName, nil
}

func isValid(jsonld string) error {
	proc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")
	options.Format = "application/nquads"

	var myInterface interface{}
	err := json.Unmarshal([]byte(jsonld), &myInterface)
	if err != nil {
		log.Println("Error when transforming JSON-LD document to interface:", err)
		return err
	}

	_, err = proc.ToRDF(myInterface, options) // returns triples but toss them, just validating
	if err != nil {
		log.Println("Error when transforming JSON-LD document to RDF:", err)
		return err
	}

	return err
}

func DomainListJSON(f string) ([]Sources, []Sources, error) {
	log.Printf("Opening source list file: %s \n", f)

	var domains []Sources

	file, err := os.Open(f)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	jp := json.NewDecoder(file)
	jp.Decode(&domains)

	hd := make([]Sources, len(domains))
	copy(hd, domains) // make sure to make with len to have "len" to copy into

	for i := len(hd) - 1; i >= 0; i-- {
		haveToDelete := hd[i].Headless
		if !haveToDelete {
			hd = append(hd[:i], hd[i+1:]...)
		}
	}

	// NOTE use downward loop to avoid removing and altering sliec index in the process
	for i := len(domains) - 1; i >= 0; i-- {
		haveToDelete := domains[i].Headless
		if haveToDelete {
			domains = append(domains[:i], domains[i+1:]...)
		}
	}

	return domains, hd, nil
}

// ResourceURLsJSON
func ResourceURLsJSON(domains []Sources) map[string]sitemaps.URLSet {
	m := make(map[string]sitemaps.URLSet) // make a map

	for k := range domains {
		if domains[k].Active {
			log.Printf("Working with active domain %s\n", domains[k].URL)
			mapname, _, err := utils.DomainNameShort(domains[k].URL)
			if err != nil {
				log.Println("Error in domain parsing")
			}
			log.Println(mapname)
			var us sitemaps.URLSet
			if domains[k].SitemapFormat == "xml" {
				us = sitemaps.IngestSitemapXML(domains[k].URL)
			}
			if domains[k].SitemapFormat == "text" {
				us = sitemaps.IngestSiteMapText(domains[k].URL)
			}
			m[mapname] = us
		}
	}
	return m
}

// ResourceURLs
func ResourceURLs(domains []string) map[string]sitemaps.URLSet {
	m := make(map[string]sitemaps.URLSet) // make a map

	for key := range domains {
		mapname, _, err := utils.DomainNameShort(domains[key])
		if err != nil {
			log.Println("Error in domain parsing")
		}
		log.Println(mapname)
		us := sitemaps.IngestSitemapXML(domains[key])
		m[mapname] = us
	}

	return m
}

// DomainList  DEPRECATED
func DomainList(f string) ([]string, error) { // return map(string[]string)
	log.Printf("Opening source list file: %s \n", f)

	var domains []string

	file, err := os.Open(f)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		domains = append(domains, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return domains, nil
}
