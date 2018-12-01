package acquire

import (
	"encoding/json"
	"log"

	"earthcube.org/Project418/gleaner/internal/summoner/sitemaps"
	"earthcube.org/Project418/gleaner/internal/utils"
	"github.com/kazarena/json-gold/ld"
)

// Sources is a struct holding the metadata associated with the sites to harvest
type Sources struct {
	Name          string
	URL           string
	Headless      bool
	SitemapFormat string
	Active        bool
}

func isValid(jsonld string) error {
	proc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")
	options.Format = "application/nquads"

	var myInterface interface{}
	err := json.Unmarshal([]byte(jsonld), &myInterface)
	if err != nil {
		log.Println("Error when transforming JSON-LD document to interface:", err)
		return err
	}

	_, err = proc.ToRDF(myInterface, options) // returns triples but toss them, just validating
	if err != nil {
		log.Println("Error when transforming JSON-LD document to RDF:", err)
		return err
	}

	return err
}

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

// ResourceURLs
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

// // ResourceURLs
// func ResourceURLs(domains []string) map[string]sitemaps.URLSet {
// 	m := make(map[string]sitemaps.URLSet) // make a map

// 	for key := range domains {
// 		mapname, _, err := utils.DomainNameShort(domains[key])
// 		if err != nil {
// 			log.Println("Error in domain parsing")
// 		}
// 		log.Println(mapname)
// 		us := sitemaps.IngestSitemapXML(domains[key])
// 		m[mapname] = us
// 	}

// 	return m
// }

// DomainList  DEPRECATED
// func DomainList(f string) ([]string, error) { // return map(string[]string)
// 	log.Printf("Opening source list file: %s \n", f)

// 	var domains []string

// 	file, err := os.Open(f)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer file.Close()

// 	scanner := bufio.NewScanner(file)
// 	for scanner.Scan() {
// 		domains = append(domains, scanner.Text())
// 	}

// 	if err := scanner.Err(); err != nil {
// 		log.Fatal(err)
// 	}

// 	return domains, nil
// }
