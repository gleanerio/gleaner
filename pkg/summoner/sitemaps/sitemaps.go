package sitemaps

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"earthcube.org/Project418/gleaner/pkg/utils"
	minio "github.com/minio/minio-go"
)

// URLSet takes a URL to a sitemap and parses out the content.
// The entries are digested one by one (later N conncurrent)
type URLSet struct {
	XMLName xml.Name  `xml:"urlset"`
	URL     []URLNode `xml:"url"`
}

// URLNode sub node struct
type URLNode struct {
	XMLName     xml.Name `xml:"url"`
	Loc         string   `xml:"loc"`
	Description string   `xml:"description"`
}

// SmIndex is the sitemap index
type SmIndex struct {
	XMLName xml.Name       `xml:"sitemapindex"`
	SiteMap []SiteMapEntry `xml:"sitemap"`
}

// SiteMapEntry are the various loc and lastmod entries on a sitemap
type SiteMapEntry struct {
	Loc     string `xml:"loc"`
	Lastmod string `xml:"lastmod"`
}

// IngestSitemapXML validates the XMl format of the sitemap and
// reads each entry into a struct array that is sent back
func IngestSitemapXML(url string, cs utils.Config) URLSet {
	// First check our URL..   if it is S3, call our S3 reader function
	var us URLSet

	if strings.HasPrefix(url, "s3://") {
		bodyBytes, err := getS3Body(url, cs) // TODO handle this error
		if err == nil {
			xml.Unmarshal(bodyBytes, &us)
		}
		return us
	}

	b, s := isSiteMapIndex(url)

	// b == false means our URL is not a sitemap, try it now as a URL node set
	if b == false {
		bodyBytes, err := getBody(url) // TODO handle this error
		if err == nil {
			xml.Unmarshal(bodyBytes, &us)
		}
	}

	una := []URLNode{}
	if b == true {
		for item := range s {
			var sm URLSet
			bodyBytes, err := getBody(s[item]) // TODO handle this error
			if err == nil {
				xml.Unmarshal(bodyBytes, &sm)
				for i := range sm.URL {
					un := URLNode{}
					un.Loc = sm.URL[i].Loc
					un.Description = sm.URL[i].Description
					una = append(una, un)
				}
			}
		}
		us.URL = una
	}

	return us
}

func getS3Body(url string, cs utils.Config) ([]byte, error) {
	// split s3://bucket/object

	mc := utils.MinioConnection(cs)

	bucket := "gleaner"
	object := "cdfsitemap.xml"

	fo, err := mc.GetObject(bucket, object, minio.GetObjectOptions{})
	if err != nil {
		fmt.Println(err)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(fo)

	return buf.Bytes(), err
}

func isSiteMapIndex(url string) (bool, []string) {
	bodyBytes, _ := getBody(url)

	var smi SmIndex
	err := xml.Unmarshal(bodyBytes, &smi)
	if err != nil {
		return false, nil
	}

	var sma []string
	for k := range smi.SiteMap {
		sma = append(sma, smi.SiteMap[k].Loc)
	}

	return true, sma
}

// IngestSiteMapText takes the URL and pulls the the URL from it
// using concept of type sitemap
func IngestSiteMapText(url string, cs utils.Config) URLSet {
	bodyBytes, _ := getBody(url) // TODO handle this error
	var sitemap URLSet

	sc := bufio.NewScanner(strings.NewReader(string(bodyBytes)))
	for sc.Scan() {
		u := sc.Text()
		un := URLNode{Loc: u}
		sitemap.URL = append(sitemap.URL, un)
	}
	if err := sc.Err(); err != nil {
		log.Fatalf("scan file error: %v", err)
	}

	return sitemap
}

func getBody(url string) ([]byte, error) {
	var client http.Client
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Print(err) // not even being able to make a req instance..  might be a fatal thing?
		return nil, err
	}

	req.Header.Set("User-Agent", "EarthCube_DataBot/1.0")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error reading sitemap: %s", err)
		return nil, err
	}
	defer resp.Body.Close()

	// var bodyString string
	var bodyBytes []byte
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Print(err)
			return nil, err
		}
	}

	return bodyBytes, err
}
