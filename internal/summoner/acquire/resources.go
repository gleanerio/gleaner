package acquire

import (
	"log"
	"net/url"
	"strings"

	"github.com/earthcubearchitecture-project418/gleaner/pkg/summoner/sitemaps"
	"github.com/spf13/viper"
)

// Sources Holds the metadata associated with the sites to harvest
type Sources struct {
	Name       string
	Logo       string
	URL        string
	Headless   bool
	PID        string
	ProperName string
	Domain     string
	// SitemapFormat string
	// Active        bool
}

// ResourceURLs looks gets the resource URLs for a domain.  The results is a
// map with domain name as key and []string of the URLs to process.
func ResourceURLs(v1 *viper.Viper, headless bool) map[string]sitemaps.Sitemap {
	m := make(map[string]sitemaps.Sitemap) // make a map

	var domains []Sources
	err := v1.UnmarshalKey("sources", &domains)
	if err != nil {
		log.Println(err)
	}

	mcfg := v1.GetStringMapString("summoner")

	for k := range domains {
		if headless == domains[k].Headless {
			log.Printf("Parsing sitemap: %s\n", domains[k].URL)
			// mapname, _, err := domainNameShort(domains[k].URL)
			mapname := domains[k].Name // TODO I would like to use this....
			if err != nil {
				log.Println("Error in domain parsing")
			}
			// These are the two lines that change in this branch
			// var us sitemaps.URLSet
			// us = sitemaps.IngestSitemap(domains[k].URL)

			log.Println(mcfg)

			var us sitemaps.Sitemap
			if mcfg["after"] != "" {
				log.Println("Get After Date")
				us, err = sitemaps.GetAfterDate(domains[k].URL, nil, mcfg["after"])
				if err != nil {
					log.Println(err)
					// pass back error and deal with it better in the logs
					// and in the user experience
				}
			} else {
				log.Println("Get with no date")
				us, err = sitemaps.Get(domains[k].URL, nil)
				if err != nil {
					log.Println(err)
					// pass back error and deal with it better in the logs
					// and in the user experience
				}
			}

			// Need to prune the us.URL array against the prov graph
			// If we request "delta run"  How to do that with JSON-LD object, not
			// a prov graph (which would need a triplestore)  ?  use golang SPARQL?

			log.Printf("%s : %d\n", domains[k].Name, len(us.URL))

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
