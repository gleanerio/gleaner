package sitemaps

// NOTE:  this code cam from https://github.com/yterajima/go-sitemap
// I copied it here to test and see if need to make some mods.
// I hope to either simply call that package as in import or
// contribute back any needed changes and then link.

import (
	"encoding/xml"
	"log"
	"strings"

	sitemap "github.com/oxffaa/gopher-parse-sitemap"
)

// Index is a structure of <sitemapindex>
type Index struct {
	XMLName xml.Name `xml:"sitemapindex"`
	Sitemap []parts  `xml:"sitemap"`
}

// parts is a structure of <sitemap> in <sitemapindex>
type parts struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod"`
}

// Sitemap is a structure of <sitemap>
type Sitemap struct {
	XMLName xml.Name `xml:":urlset"`
	URL     []URL    `xml:":url"`
}

// URL is a structure of <url> in <sitemap>
type URL struct {
	Loc        string  `xml:"loc"`
	LastMod    string  `xml:"lastmod"`
	ChangeFreq string  `xml:"changefreq"`
	Priority   float32 `xml:"priority"`
}

// returns an array of sitemaps from sitemap index
func DomainIndex(sm string) ([]string, error) {
	result := make([]string, 0)
	err := sitemap.ParseIndexFromSite(sm, func(e sitemap.IndexEntry) error {
		result = append(result, strings.TrimSpace(e.GetLocation()))
		return nil
	})

	return result, err
}

// returns a Sitemap struct from a sitemap
func DomainSitemap(sm string) (Sitemap, error) {
	// result := make([]string, 0)
	smsm := Sitemap{}

	urls := make([]URL, 0)
	err := sitemap.ParseFromSite(sm, func(e sitemap.Entry) error {
		entry := URL{}
		entry.Loc = strings.TrimSpace(e.GetLocation())
		//TODO why is this failing?  The string doesn't exist..  need to test and trap
		// 	entry.LastMod = e.GetLastModified().String()
		// entry.ChangeFreq = strings.TrimSpace(e.GetChangeFrequency())
		urls = append(urls, entry)
		return nil
	})

	if err != nil {
		log.Println(err)
	}

	smsm.URL = urls
	return smsm, err
}
