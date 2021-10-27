package acquire

import (
	"fmt"
	"log"
	"strings"

	configTypes "github.com/gleanerio/gleaner/internal/config"

	"github.com/boltdb/bolt"
	"github.com/gleanerio/gleaner/internal/objects"
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

// ResourceURLs looks gets the resource URLs for a domain.  The results is a
// map with domain name as key and []string of the URLs to process.
func ResourceURLs(v1 *viper.Viper, mc *minio.Client, headless bool, db *bolt.DB) map[string][]string {
	// read config file
	//miniocfg := v1.GetStringMapString("minio")
	//bucketName := miniocfg["bucket"] //   get the top level bucket for all of gleaner operations from config file

	m := make(map[string][]string) // make a map

	//var domains []Sources
	//err := v1.UnmarshalKey("sources", &domains)
	domains, err := configTypes.ParseSourcesConfig(v1)
	domains = configTypes.GetSourceByType(domains, siteMapType)
	if err != nil {
		log.Println(err)
	}

	mcfg := v1.GetStringMapString("summoner")

	for k := range domains {
		if headless == domains[k].Headless {
			mapname := domains[k].Name // TODO I would like to use this....
			if err != nil {
				log.Println(domains[k].Name, "Error in domain parsing", err)
			}

			// log.Println(mcfg)

			var us sitemaps.Sitemap

			idxr, err := sitemaps.DomainIndex(domains[k].URL)
			if err != nil {
				log.Println("Error reading this source")
				log.Println(err)
				// os.Exit(0)  // TODO this function needs to return an error and talk with others
			}

			if len(idxr) < 1 {
				log.Println("We are not a sitemap index, check to see if we are a sitemap")
				us, err = sitemaps.DomainSitemap(domains[k].URL)
				if err != nil {
					fmt.Println(err)
				}
			} else {
				log.Println("Walk the sitemap index for sitemaps")
				for _, idxv := range idxr {
					subset, err := sitemaps.DomainSitemap(idxv)
					us.URL = append(us.URL, subset.URL...)
					if err != nil {
						fmt.Println(err)
					}
				}
			}

			// if mcfg["after"] != "" {
			// 	//log.Println("Get After Date")
			// 	us, err = sitemaps.GetAfterDate(domains[k].URL, nil, mcfg["after"])
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
			var s []string
			for k := range us.URL {
				if us.URL[k].Loc != "" { // TODO why did this otherwise add a nil to the array..  need to check
					s = append(s, strings.TrimSpace(us.URL[k].Loc))
				}
			}

			//  TODO if we check for URLs in prov..  do that here..
			if mcfg["mode"] == "diff" {
				//oa := objects.ProvURLs(v1, mc, bucketName, fmt.Sprintf("prov/%s", mapname))

				oa := []string{}
				db.View(func(tx *bolt.Tx) error {
					// Assume bucket exists and has keys
					b := tx.Bucket([]byte(domains[k].Name))
					c := b.Cursor()

					for key, _ := c.First(); key != nil; key, _ = c.Next() {
						//fmt.Printf("key=%s, value=%s\n", k, v)
						oa = append(oa, fmt.Sprintf("%s", key))
					}

					return nil
				})

				d := difference(s, oa)
				m[mapname] = d
			} else {
				m[mapname] = s
			}

			//  TODO if we check for URLs in prov..  do that here..
			//if mcfg["mode"] == "diff" {
			//oa := objects.ProvURLs(v1, mc, bucketName, fmt.Sprintf("prov/%s", mapname))
			//d := difference(s, oa)
			//m[mapname] = d
			//} else {
			//m[mapname] = s
			//}

			log.Printf("%s sitemap size is : %d queuing: %d mode: %s \n", domains[k].Name, len(s), len(m[mapname]), mcfg["mode"])

		}
	}

	return m
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
