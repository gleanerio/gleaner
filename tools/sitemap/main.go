package main

import (
	"fmt"
	"os"
	"strings"

	// "github.com/gleanerio/gleaner/internal/summoner/sitemaps"
	"github.com/gleanerio/gleaner/internal/summoner/sitemaps"
	sitemap "github.com/oxffaa/gopher-parse-sitemap"
)

func main() {

	source := "https://geoconnex.us/sitemap.xml"
	// source := "https://geoconnex.us/sitemap/namespaces/CHyLD/chyld-pilot_ids__0.xml"
	// source := "https://samples.earth/sitemap.xml"

	// Look for robots.txt
	if strings.HasSuffix(source, "robots.txt") {
		fmt.Println("robots.txt support coming soon")
		os.Exit(0)
	}

	idxr, err := DomainIndex(source)
	if err != nil {
		fmt.Println("Error reading this source")
		fmt.Println(err)
		os.Exit(0)
	}

	if len(idxr) < 1 {
		fmt.Println("We are not a sitemap index, check to see if we are a sitemap")
		sr, _ := DomainSitemap(source)
		for _, v := range sr.URL {
			fmt.Println(v)
		}
	} else {
		for _, idxv := range idxr {
			fmt.Println(idxv)
			//smr, err := DomainSitemap(idxv)
			//if err != nil {
			//fmt.Println(err)
			//}
			//for _, smv := range smr.URL {
			//fmt.Println(smv.Loc)
			//}
		}
	}

}

func DomainSitemap(sm string) (sitemaps.Sitemap, error) {
	// result := make([]string, 0)
	smsm := sitemaps.Sitemap{}
	urls := make([]sitemaps.URL, 0)
	err := sitemap.ParseFromSite(sm, func(e sitemap.Entry) error {
		entry := sitemaps.URL{}
		entry.Loc = strings.TrimSpace(e.GetLocation())
		//TODO why is this failing?  The string doesn't exist..  need to test and trap
		// 	entry.LastMod = e.GetLastModified().String()
		// entry.ChangeFreq = strings.TrimSpace(e.GetChangeFrequency())
		urls = append(urls, entry)
		return nil
	})

	smsm.URL = urls
	return smsm, err
}

func DomainIndex(sm string) ([]string, error) {
	result := make([]string, 0)
	err := sitemap.ParseIndexFromSite(sm, func(e sitemap.IndexEntry) error {
		result = append(result, strings.TrimSpace(e.GetLocation()))
		return nil
	})

	return result, err
}
