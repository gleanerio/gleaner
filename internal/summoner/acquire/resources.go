package acquire

import (
	"fmt"
	"github.com/gleanerio/gleaner/internal/common"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"

	configTypes "github.com/gleanerio/gleaner/internal/config"
	bolt "go.etcd.io/bbolt"

	"github.com/temoto/robotstxt"

	"github.com/gleanerio/gleaner/internal/summoner/sitemaps"
	"github.com/minio/minio-go/v7"
	"github.com/spf13/viper"
)

// Sources Holds the metadata associated with the sites to harvest
//
//	type Sources struct {
//		Name       string
//		Logo       string
//		URL        string
//		Headless   bool
//		PID        string
//		ProperName string
//		Domain     string
//		// SitemapFormat string
//		// Active        bool
//	}
//
// type Sources = configTypes.Sources
const siteMapType = "sitemap"
const robotsType = "robots"

// ResourceURLs looks gets the resource URLs for a domain.  The results is a
// map with domain name as key and []string of the URLs to process.
func ResourceURLs(v1 *viper.Viper, mc *minio.Client, headless bool, db *bolt.DB) (map[string][]string, error) {
	domainsMap := make(map[string][]string)
	var repoFatalErrors common.MultiError
	// Know whether we are running in diff mode, in order to exclude urls that have already
	// been summoned before
	mcfg, err := configTypes.ReadSummmonerConfig(v1.Sub("summoner"))
	sources, err := configTypes.GetSources(v1)
	domains := configTypes.GetActiveSourceByHeadless(sources, headless)
	if err != nil {
		log.Error("Error getting sources to summon: ", err)
		return domainsMap, err // if we can't read list, ok to return an error
	}

	sitemapDomains := configTypes.GetActiveSourceByType(domains, siteMapType)

	for _, domain := range sitemapDomains {
		var robots *robotstxt.RobotsData
		var group *robotstxt.Group

		if v1.Get("rude") == true {
			robots = nil
			group = nil
			log.Info("Rude indexing mode enabled; ignoring robots.txt.")
		} else {
			robots, err = getRobotsForDomain(domain.Domain)
			if err != nil {
				log.Error("Error getting robots.txt for ", domain.Name, ", continuing without it.")
				robots = nil
				group = nil
			}
		}
		if robots != nil {
			group = robots.FindGroup(EarthCubeAgent)
			log.Info("Got robots.txt group ", group)
		}
		urls, err := getSitemapURLList(domain.URL, group)
		if err != nil {
			log.Error("Error getting sitemap urls for: ", domain.Name, err)
			repoFatalErrors = append(repoFatalErrors, err)
			//return domainsMap, err // returning means that domains after broken one do not get indexed.
		}
		if mcfg.Mode == "diff" {
			urls = excludeAlreadySummoned(domain.Name, urls, db)
		}
		overrideCrawlDelayFromRobots(v1, domain.Name, mcfg.Delay, group)
		domainsMap[domain.Name] = urls
		log.Debug(domain.Name, "sitemap size is :", len(domainsMap[domain.Name]), " mode: ", mcfg.Mode)
	}

	robotsDomains := configTypes.GetActiveSourceByType(domains, robotsType)

	for _, domain := range robotsDomains {

		var urls []string
		// first, get the robots file and parse it
		robots, err := getRobotsTxt(domain.URL)
		if err != nil {
			log.Error("Error getting sitemap location from robots.txt for: ", domain.Name, err)
			repoFatalErrors = append(repoFatalErrors, err)
			//return domainsMap, err // returning means that domains after broken one do not get indexed.
		}
		group := robots.FindGroup(EarthCubeAgent)
		log.Debug("Found user agent group ", group)
		for _, sitemap := range robots.Sitemaps {
			sitemapUrls, err := getSitemapURLList(sitemap, group)
			if err != nil {
				log.Error("Error getting sitemap urls for: ", domain.Name, err)
				repoFatalErrors = append(repoFatalErrors, err)
				//return domainsMap, err // returning means that domains after broken one do not get indexed.
			}
			urls = append(urls, sitemapUrls...)
		}
		if mcfg.Mode == "diff" {
			urls = excludeAlreadySummoned(domain.Name, urls, db)
		}
		overrideCrawlDelayFromRobots(v1, domain.Name, mcfg.Delay, group)
		domainsMap[domain.Name] = urls
		log.Debug(domain.Name, "sitemap size from robots.txt is : ", len(domainsMap[domain.Name]), " mode: ", mcfg.Mode)
	}
	if len(repoFatalErrors) == 0 {
		return domainsMap, nil
	} else {
		return domainsMap, repoFatalErrors
	}

}

// given a sitemap url, parse it and get the list of URLS from it.
func getSitemapURLList(domainURL string, robots *robotstxt.Group) ([]string, error) {
	var us sitemaps.Sitemap
	var s []string

	idxr, err := sitemaps.DomainIndex(domainURL)
	if err != nil {
		log.Error("Error reading sitemap at:", domainURL, err)
		return s, err
	}

	if len(idxr) < 1 {
		log.Info(domainURL, " is not a sitemap index, checking to see if it is a sitemap")
		us, err = sitemaps.DomainSitemap(domainURL)
		if err != nil {
			log.Error("Error parsing sitemap index at ", domainURL, err)
			return s, err
		}
	} else {
		log.Info("Walking the sitemap index for sitemaps")
		for _, idxv := range idxr {
			subset, err := sitemaps.DomainSitemap(idxv)
			us.URL = append(us.URL, subset.URL...)
			if err != nil {
				log.Error("Error parsing sitemap index at: ", idxv, err)
				return s, err
			}
		}
	}

	// Convert the array of sitemap package struct to simply the URLs in []string
	for _, urlStruct := range us.URL {
		if urlStruct.Loc != "" { // TODO why did this otherwise add a nil to the array..  need to check
			loc := strings.TrimSpace(urlStruct.Loc)
			loc = strings.ReplaceAll(loc, " ", "")
			loc = strings.ReplaceAll(loc, "\n", "")

			if robots != nil && !robots.Test(loc) {
				log.Error("Declining to index ", loc, " because it is disallowed by robots.txt. Error information, if any:", err)
				continue
			}
			s = append(s, loc)
		}
	}

	return s, nil
}

func overrideCrawlDelayFromRobots(v1 *viper.Viper, sourceName string, delay int64, robots *robotstxt.Group) {
	// Look at the crawl delay from this domain's robots.txt, if we can, and one exists.
	if robots != nil {
		// this is a time.Duration, which is in nanoseconds, because of COURSE it is, but we want milliseconds
		log.Debug("Raw crawl delay for robots ", sourceName, " set to ", robots.CrawlDelay)
		crawlDelay := int64(robots.CrawlDelay / time.Millisecond)
		log.Debug("Crawl Delay specified by robots.txt for ", sourceName, " : ", crawlDelay)

		// If our default delay is less than what is set there, set a delay for this
		// domain to respect the robots.txt setting.
		if delay < crawlDelay {
			sources, err := configTypes.GetSources(v1)
			source, err := configTypes.GetSourceByName(sources, sourceName)

			if err != nil {
				log.Error("Error setting crawl delay override for ", sourceName, ":", err)
				return
			}
			source.Delay = crawlDelay
			v1.Set("sources", sources)
		}
	}
}

func getRobotsForDomain(url string) (*robotstxt.RobotsData, error) {
	robotsUrl := url + "/robots.txt"
	log.Info("Getting robots.txt from ", robotsUrl)
	robots, err := getRobotsTxt(robotsUrl)
	if err != nil {
		log.Error("error getting robots.txt for ", url, ":", err)
		return nil, err
	}
	return robots, nil
}

func excludeAlreadySummoned(domainName string, urls []string, db *bolt.DB) []string {
	//  TODO if we check for URLs in prov..  do that here..
	//oa := objects.ProvURLs(v1, mc, bucketName, fmt.Sprintf("prov/%s", domain.Name))

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
