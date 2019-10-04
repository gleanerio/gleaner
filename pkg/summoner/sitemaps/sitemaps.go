package sitemaps

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
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

// IngestSitemap validates the XMl format of the sitemap and
// reads each entry into a struct array that is sent back
func IngestSitemap(url string) URLSet {
	// get the body then check the type
	// if text, use the text code else..  use the XML code (do a case switch here perhaps?)

	bb, err := getBody(url)
	if err != nil {
		log.Println(err)
	}

	// I would like to be able to use this to parse between XML and Text versions of the sitemap
	ct := http.DetectContentType(bb)
	log.Printf("Content type of sitemap reference is %s\n", ct)

	var us URLSet
	una := []URLNode{}

	b, s := isSiteMapIndex(bb)

	if b == false { // not a sitemap
		xml.Unmarshal(bb, &us)
	} else { // is a sitemap
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

func isSiteMapIndex(bodyBytes []byte) (bool, []string) {
	var smi SmIndex
	err := xml.Unmarshal(bodyBytes, &smi)
	if err != nil {
		return false, nil
	}

	// We seem to be a sitemap, so let's try and parse it..
	var sma []string
	for k := range smi.SiteMap {
		sma = append(sma, smi.SiteMap[k].Loc)
	}

	return true, sma
}

// // IngestSiteMapText takes the URL and pulls the the URL from it
// func IngestSiteMapText(url string, cs utils.Config) URLSet {
// 	bodyBytes, _ := getBody(url) // TODO handle this error
// 	var us URLSet

// 	sc := bufio.NewScanner(strings.NewReader(string(bodyBytes)))
// 	for sc.Scan() {
// 		u := sc.Text()
// 		un := URLNode{Loc: u}
// 		us.URL = append(us.URL, un)
// 	}
// 	if err := sc.Err(); err != nil {
// 		log.Fatalf("scan file error: %v", err)
// 	}

// 	return us
// }

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
