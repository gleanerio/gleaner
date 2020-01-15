package acquire

import (
	"log"
	"net/url"
	"strings"

	"earthcube.org/Project418/gleaner/pkg/summoner/sitemaps"
	"github.com/spf13/viper"
)

// Sources Holds the metadata associated with the sites to harvest
type Sources struct {
	Name     string
	Logo     string
	URL      string
	Headless bool
	// SitemapFormat string
	// Active        bool
}

// ResourceURLs looks gets the resource URLs for a domain.  The results is a
// map with domain name as key and []string of the URLs to process.
func ResourceURLs(v1 *viper.Viper, headless bool) map[string]sitemaps.URLSet {
	m := make(map[string]sitemaps.URLSet) // make a map

	var domains []Sources
	err := v1.UnmarshalKey("sources", &domains)
	if err != nil {
		log.Println(err)
	}

	for k := range domains {
		if headless == domains[k].Headless {
			log.Printf("Parsing sitemap: %s\n", domains[k].URL)
			// mapname, _, err := domainNameShort(domains[k].URL)
			mapname := domains[k].Name // TODO I would like to use this....
			if err != nil {
				log.Println("Error in domain parsing")
			}
			var us sitemaps.URLSet
			us = sitemaps.IngestSitemap(domains[k].URL)
			m[mapname] = us
		}
	}

	return m
}

func domainNameShort(dn string) (string, string, error) {
	u, err := url.Parse(dn)
	if err != nil {
		log.Printf("Error with domainNameShort: %s ;  %s", dn, err)
	}

	return strings.Replace(u.Host, ".", "", -1), u.Scheme, err
}
