package acquire

import (
	"fmt"
	"log"

	"earthcube.org/Project418/gleaner/pkg/summoner/sitemaps"
	"earthcube.org/Project418/gleaner/pkg/utils"
	"github.com/minio/minio-go"
)

// Sources Holds the metadata associated with the sites to harvest
type Sources struct {
	Name          string
	URL           string
	Headless      bool
	SitemapFormat string
	Active        bool
}

// DomainListJSON gets all the domains we will be working with.  It will
// return two arrays (domains, and domains that need headless processing) along
// any error that might occur.
func DomainListJSON(cs utils.Config) ([]Sources, []Sources, error) {
	// log.Printf("Opening source list file: %s \n", f)

	var domains []Sources

	for _, v := range cs.Sources {
		source := Sources{Name: v.Name, URL: v.URL, Headless: v.Headless,
			SitemapFormat: v.Sitemapformat, Active: v.Active}
		domains = append(domains, source)
	}

	hd := make([]Sources, len(domains))
	copy(hd, domains) // make sure to make with len to have "len" to copy into

	for i := len(hd) - 1; i >= 0; i-- {
		haveToDelete := hd[i].Headless
		if !haveToDelete {
			hd = append(hd[:i], hd[i+1:]...)
		}
	}

	// NOTE use downward loop to avoid removing and altering slice index in the process
	for i := len(domains) - 1; i >= 0; i-- {
		haveToDelete := domains[i].Headless
		if haveToDelete {
			domains = append(domains[:i], domains[i+1:]...)
		}
	}

	return domains, hd, nil
}

// ResourceURLs looks gets the resource URLs for a domain.  The results is a
// map with domain name as key and []string of the URLs to process.
func ResourceURLs(domains []Sources, cs utils.Config) map[string]sitemaps.URLSet {
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
				us = sitemaps.IngestSitemapXML(domains[k].URL, cs)
			}
			if domains[k].SitemapFormat == "text" {
				us = sitemaps.IngestSiteMapText(domains[k].URL, cs)
			}
			m[mapname] = us
		}
	}
	return m
}

// TODO if we default to glaner-summoned then no buckets are needed since we work with prefixes..
// buildBuckets generates the needed buckets for a run.
func buildBuckets(minioClient *minio.Client, m map[string]sitemaps.URLSet) error {
	log.Println("Building buckets")
	var err error

	for k := range m {
		// bucketName := fmt.Sprintf("gleaner-summoned/%s", k) // old was just k
		bucketName := "gleaner-summoned" // old was just k
		fmt.Printf("Keeping k alive during testing %s\n", k)
		location := "us-east-1"
		err = minioClient.MakeBucket(bucketName, location)
		if err != nil {
			// Check to see if we already own this bucket (which happens if you run this twice)
			exists, err := minioClient.BucketExists(bucketName)
			if err == nil && exists {
				log.Printf("We already own %s, deleting current objects\n", bucketName)
			} else {
				log.Fatalln(err)
			}

			// TODO   should I empty the bucket if it exists?  (make this a flag?)
			objectsCh := make(chan string)
			// Send object names that are needed to be removed to objectsCh
			go func() {
				defer close(objectsCh)
				// List all objects from a bucket-name with a matching prefix.
				for object := range minioClient.ListObjects(bucketName, "", true, nil) {
					if object.Err != nil {
						log.Fatalln(object.Err)
					}
					objectsCh <- object.Key
				}
			}()

			// NOTE removed the delete...  need to resolve this better with node ModDate and
			// the existence of prefixes!

			// TODO  Update or remove this?  Now we use prefixes we don't remove gleaner-summoned ..
			// but if we could remove by prefix that would fine
			// for rErr := range minioClient.RemoveObjects(bucketName, objectsCh) {
			// 	fmt.Println("Error detected during deletion: ", rErr)
			// }
		}
		log.Printf("Successfully created %s\n", bucketName)
	}

	return err
}
