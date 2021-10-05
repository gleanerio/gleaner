package acquire

import (
	"fmt"
	"log"

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

// ResourceURLs looks gets the resource URLs for a domain.  The results is a
// map with domain name as key and []string of the URLs to process.
func ResourceURLs(v1 *viper.Viper, mc *minio.Client, headless bool) map[string][]string {
	// read config file
	miniocfg := v1.GetStringMapString("minio")
	bucketName := miniocfg["bucket"] //   get the top level bucket for all of gleaner operations from config file

	m := make(map[string][]string) // make a map

	var domains []objects.Sources
	err := v1.UnmarshalKey("sources", &domains)
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
			if mcfg["after"] != "" {
				//log.Println("Get After Date")
				us, err = sitemaps.GetAfterDate(domains[k].URL, nil, mcfg["after"])
				if err != nil {
					log.Println(domains[k].Name, err)
					// pass back error and deal with it better in the logs
					// and in the user experience
				}
			} else {
				//log.Println("Get with no date")
				us, err = sitemaps.Get(domains[k].URL, nil)
				if err != nil {
					log.Println(domains[k].Name, err)
					// pass back error and deal with it better in the logs
					// and in the user experience
				}
			}

			// Convert the array of sitemap package stuct to simply the URLs in []string
			var s []string
			for k := range us.URL {
				if us.URL[k].Loc != "" { // TODO why did this otherwise add a nil to the array..  ned to check
					s = append(s, us.URL[k].Loc)
				}
			}
			log.Printf("crawling %s : %d\n urls", domains[k].Name, len(s))

			// TODO if we check for URLs in prov..  do that here..
			if mcfg["mode"] == "diff" {
				log.Println("doing a diff call")
				oa := objects.ProvURLs(v1, mc, bucketName, fmt.Sprintf("prov/%s", mapname))

				d := difference(s, oa)

				m[mapname] = d
			} else {
				m[mapname] = s
			}
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
