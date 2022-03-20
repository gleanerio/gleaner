package acquire

import (
	"fmt"
	"log"
	"strings"

	configTypes "github.com/gleanerio/gleaner/internal/config"
	bolt "go.etcd.io/bbolt"

	"github.com/gleanerio/gleaner/internal/summoner/sitemaps"
	"github.com/minio/minio-go/v7"
	"github.com/spf13/viper"
)

// Sources Holds the metadata associated with the sites to harvest
// type Sources struct {
// 	Name       string
// 	Logo       string
// 	URL        string
// 	Headless   bool
// 	PID        string
// 	ProperName string
// 	Domain     string
// 	// SitemapFormat string
// 	// Active        bool
// }
//type Sources = configTypes.Sources
const siteMapType = "sitemap"
const robotsType = "robots"

// ResourceURLs looks gets the resource URLs for a domain.  The results is a
// map with domain name as key and []string of the URLs to process.
func ResourceURLs(v1 *viper.Viper, mc *minio.Client, headless bool, db *bolt.DB) (map[string][]string, error) {
	m := make(map[string][]string) // make a map

	// Know whether we are running in diff mode, in order to exclude urls that have already
	// been summoned before
	mcfg, err := configTypes.ReadSummmonerConfig(v1.Sub("summoner"))
	domains, err := configTypes.GetSources(v1)
	domains = configTypes.GetActiveSourceByHeadless(domains, headless)
	if err != nil {
		log.Println("Error getting sources to summon: ", err)
		return m, err
	}

	sitemapDomains := configTypes.GetActiveSourceByType(domains, siteMapType)

	for _, domain := range sitemapDomains {
		mapname := domain.Name
		urls, err := getSitemapURLList(domain.URL)
		if err != nil {
			log.Println("Error getting sitemap urls for: ", mapname, err)
			return m, err
		}
		if mcfg.Mode == "diff" {
			urls = excludeAlreadySummoned(mapname, urls, db)
		}
		m[mapname] = urls
		log.Printf("%s sitemap size is : %d mode: %s \n", mapname, len(m[mapname]), mcfg.Mode)
	}

	robotsDomains := configTypes.GetActiveSourceByType(domains, robotsType)

	for _, domain := range robotsDomains {
		mapname := domain.Name
		var urls []string
		// first, get the robots file and parse it
		robots, err := getRobotsTxt(domain.URL)
		if err != nil {
			log.Println("Error getting sitemap location from robots.txt for: ", mapname, err)
			return m, err
		}
		for _, sitemap := range robots.Sitemaps() {
			sitemapUrls, err := getSitemapURLList(sitemap)
			if err != nil {
				log.Println("Error getting sitemap urls for: ", mapname, err)
				return m, err
			}
			urls = append(urls, sitemapUrls...)
		}
		if mcfg.Mode == "diff" {
			urls = excludeAlreadySummoned(mapname, urls, db)
		}
		m[mapname] = urls
		log.Printf("%s sitemap size from robots.txt is : %d mode: %s \n", mapname, len(m[mapname]), mcfg.Mode)
	}

	return m, nil
}

// given a sitemap url, parse it and get the list of URLS from it.
func getSitemapURLList(domainURL string) ([]string, error) {
	var us sitemaps.Sitemap
	var s []string

	idxr, err := sitemaps.DomainIndex(domainURL)
	if err != nil {
		log.Println("Error reading sitemap at:", domainURL, err)
		return s, err
	}

	if len(idxr) < 1 {
		log.Println(domainURL, "is not a sitemap index, checking to see if it is a sitemap")
		us, err = sitemaps.DomainSitemap(domainURL)
		if err != nil {
			log.Println("Error parsing sitemap index at ", domainURL, err)
			return s, err
		}
	} else {
		log.Println("Walking the sitemap index for sitemaps")
		for _, idxv := range idxr {
			subset, err := sitemaps.DomainSitemap(idxv)
			us.URL = append(us.URL, subset.URL...)
			if err != nil {
				log.Println("Error parsing sitemap index at:", idxv, err)
				return s, err
			}
		}
	}
	// if mcfg["after"] != "" {
	// 	//log.Println("Get After Date")
	// 	us, err = sitemaps.GetAfterDate(domain.URL, nil, mcfg["after"])
	// 	if err != nil {
	// 		log.Println(domains[k].Name, err)
	// 		// pass back error and deal with it better in the logs
	// 		// and in the user experience
	// 	}
	// } else {
	// 	//log.Println("Get with no date")
	// 	us, err = sitemaps.Get(domains[k].URL, nil)
	// 	if err != nil {
	// 		log.Println(domains[k].Name, err)
	// 		// pass back error and deal with it better in the logs
	// 		// and in the user experience
	// 	}
	// }

	// Convert the array of sitemap package struct to simply the URLs in []string
	for _, urlStruct := range us.URL {
		if urlStruct.Loc != "" { // TODO why did this otherwise add a nil to the array..  need to check
			s = append(s, strings.TrimSpace(urlStruct.Loc))
		}
	}

	return s, nil
}

func excludeAlreadySummoned(domainName string, urls []string, db *bolt.DB) []string {
	//  TODO if we check for URLs in prov..  do that here..
	//oa := objects.ProvURLs(v1, mc, bucketName, fmt.Sprintf("prov/%s", mapname))

	oa := []string{}
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(domainName))
		c := b.Cursor()

		for key, _ := c.First(); key != nil; key, _ = c.Next() {
			//fmt.Printf("key=%s, value=%s\n", k, v)
			oa = append(oa, fmt.Sprintf("%s", key))
		}

		return nil
	})

	d := difference(urls, oa)
	return d
}

// difference returns the elements in `a` that aren't in `b`.
func difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}
