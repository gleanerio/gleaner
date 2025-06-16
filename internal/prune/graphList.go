package prune

import (
	"bytes"
	"fmt"
	"github.com/gleanerio/gleaner/internal/graph"
	"github.com/gleanerio/gleaner/nabu/pkg/config"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func graphList(v1 *viper.Viper, prefix string) ([]string, error) {
	log.Println("Getting list of named graphs")

	var ga []string

	ep := v1.GetString("flags.endpoint")
	spql, err := config.GetEndpoint(v1, ep, "sparql")
	if err != nil {
		log.Error(err)
	}

	//bucketName, err := config.GetBucketName(v1)
	//if err != nil {
	//	log.Println(err)
	//	return ga, err
	//}

	gp, err := graph.MakeURNPrefix(v1, prefix)
	if err != nil {
		log.Println(err)
		return ga, err
	}

	//d := fmt.Sprintf("SELECT DISTINCT ?g WHERE {GRAPH ?g {?s ?p ?o} FILTER regex(str(?g), \"^%s\")}", gp)

	d := fmt.Sprint("SELECT DISTINCT ?g WHERE {GRAPH ?g {?s ?p ?o} }")

	log.Printf("Pattern: %s\n", gp)
	log.Printf("SPARQL: %s\n", d)
	//log.Printf("Accept: %s\n", spql.Accept)
	//log.Printf("URL: %s\n", spql.URL)

	pab := []byte("")
	params := url.Values{}
	params.Add("query", d)
	//req, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", spql["endpoint"], params.Encode()), bytes.NewBuffer(pab))
	req, err := http.NewRequest(spql.Method, fmt.Sprintf("%s?%s", spql.URL, params.Encode()), bytes.NewBuffer(pab))
	if err != nil {
		log.Println(err)
	}

	// These headers
	req.Header.Set("Accept", spql.Accept)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(strings.Repeat("ERROR", 5))
		log.Println("response Status:", resp.Status)
		log.Println("response Headers:", resp.Header)
		log.Println("response Body:", string(body))
	}

	// debugging calls
	//fmt.Println("response Body:", string(body))
	//err = ioutil.WriteFile("body.txt", body, 0644)
	//if err != nil {
	//	fmt.Println("An error occurred:", err)
	//	return ga, err
	//}

	result := gjson.Get(string(body), "results.bindings.#.g.value")
	result.ForEach(func(key, value gjson.Result) bool {
		ga = append(ga, value.String())
		return true // keep iterating
	})

	var gaf []string
	for _, str := range ga {
		if strings.HasPrefix(str, gp) { // check if string has prefix
			gaf = append(gaf, str) // if yes, add it to newArray
		}
	}

	return gaf, nil
}
