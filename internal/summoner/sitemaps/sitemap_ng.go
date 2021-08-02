package sitemaps

// NOTE:  this code cam from https://github.com/yterajima/go-sitemap
// I copied it here to test and see if need to make some mods.
// I hope to either simply call that package as in import or
// contribute back any needed changes and then link.

import (
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/araddon/dateparse"
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
	XMLName xml.Name `xml:"urlset"`
	URL     []URL    `xml:"url"`
}

// URL is a structure of <url> in <sitemap>
type URL struct {
	Loc        string  `xml:"loc"`
	LastMod    string  `xml:"lastmod"`
	ChangeFreq string  `xml:"changefreq"`
	Priority   float32 `xml:"priority"`
}

// GetAfterDate sitemap data from URL returning only those after a date
func GetAfterDate(URL string, options interface{}, date string) (Sitemap, error) {
	data, err := fetch(URL, options)
	if err != nil {
		return Sitemap{}, err
	}

	idx, idxErr := ParseIndex(data)
	smap, smapErr := Parse(data)

	if idxErr != nil && smapErr != nil {
		return Sitemap{}, errors.New("URL is not a sitemap or sitemapindex")
	} else if idxErr != nil {
		return smap, nil
	}

	smap, err = idx.get(data, options)
	if err != nil {
		return Sitemap{}, err
	}

	// var c []string
	// var culled Sitemap

	if date != "" {
		for i := len(smap.URL) - 1; i >= 0; i-- {

			if smap.URL[i].LastMod != "" {
				t, err := dateparse.ParseAny(smap.URL[i].LastMod)
				if err != nil {
					log.Println(err)
				}
				check, err := time.Parse(time.RFC822, date)
				if err != nil {
					log.Println(err)
				}
				q := afterTime(t, check) // tru if t after check

				// remove the item if FALSE
				if !q {
					smap.URL = append(smap.URL[:i],
						smap.URL[i+1:]...)
				}
			}
		}
	}

	return smap, nil
}

// Get sitemap data from URL
func Get(URL string, options interface{}) (Sitemap, error) {
	data, err := fetch(URL, options)
	if err != nil {
		return Sitemap{}, err
	}

	idx, idxErr := ParseIndex(data)
	smap, smapErr := Parse(data)

	if idxErr != nil && smapErr != nil {
		log.Println(idxErr)
		log.Println(smapErr)
		return Sitemap{}, errors.New("URL is not a sitemap or sitemapindex")
	} else if idxErr != nil {
		// log.Println(idxErr)
		return smap, errors.New("URL is not a sitemapindex")
	}

	smap, err = idx.get(data, options)
	if err != nil {
		log.Println(err)
		return Sitemap{}, err
	}

	return smap, nil
}

func gUnzipData(data []byte) (resData []byte, err error) {
	b := bytes.NewBuffer(data)

	var r io.Reader
	r, err = gzip.NewReader(b)
	if err != nil {
		return
	}

	var resB bytes.Buffer
	_, err = resB.ReadFrom(r)
	if err != nil {
		return
	}

	resData = resB.Bytes()

	return
}

// aftertime returns a boolean true if check is after lastmod
func afterTime(lastmod, check time.Time) bool {
	return lastmod.After(check)
}

// fetch is page acquisition function
var fetch = func(URL string, options interface{}) ([]byte, error) {
	var body []byte

	res, err := http.Get(URL)
	if err != nil {
		return body, err
	}
	defer res.Body.Close()

	// TODO  move the gunzip here..
	// if url ends in .gz then returnthe uncompressed bytes...

	b, err := ioutil.ReadAll(res.Body)
	var data []byte
	if strings.HasSuffix(URL, ".gz") {
		// log.Println("Gziped sitemap")
		data, err = gUnzipData(b)
		if err != nil {
			log.Println(err)
		}
	} else {
		// log.Println("Uncompressed XML sitemap")
		data = append(data, b...)
	}

	return data, err
}

// Time interval to be used in Index.get
var interval = time.Second

// Get Sitemap data from sitemapindex file
func (s *Index) get(data []byte, options interface{}) (Sitemap, error) {
	idx, err := ParseIndex(data)
	if err != nil {
		return Sitemap{}, err
	}

	var smap Sitemap
	for _, s := range idx.Sitemap {
		time.Sleep(interval)
		data, err := fetch(s.Loc, options)
		if err != nil {
			return smap, err
		}

		err = xml.Unmarshal(data, &smap)
		if err != nil {
			return smap, err
		}
	}

	return smap, err
}

// Parse create Sitemap data from text
func Parse(data []byte) (smap Sitemap, err error) {
	err = xml.Unmarshal(data, &smap)
	return
}

// ParseIndex create Index data from text
func ParseIndex(data []byte) (idx Index, err error) {
	err = xml.Unmarshal(data, &idx)
	return
}

// SetInterval change Time interval to be used in Index.get
func SetInterval(time time.Duration) {
	interval = time
}

// SetFetch change fetch closure
func SetFetch(f func(URL string, options interface{}) ([]byte, error)) {
	fetch = f
}
