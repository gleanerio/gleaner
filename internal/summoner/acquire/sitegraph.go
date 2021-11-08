package acquire

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	configTypes "github.com/gleanerio/gleaner/internal/config"

	"github.com/gleanerio/gleaner/internal/common"
	"github.com/gleanerio/gleaner/internal/millers/graph"
	"github.com/minio/minio-go/v7"
	"github.com/spf13/viper"
)

type Sources = configTypes.Sources

const siteGraphType = "sitegraph"

// GetGraph downloads pre-built site graphs
func GetGraph(mc *minio.Client, v1 *viper.Viper) (string, error) {
	// read config file
	//miniocfg := v1.GetStringMapString("minio")
	bucketName, err := configTypes.GetBucketName(v1) //miniocfg["bucket"] //   get the top level bucket for all of gleaner operations from config file

	// get the sitegraph entry from config file
	var domains []Sources
	//err := v1.UnmarshalKey("sitegraphs", &domains)

	sources, err := configTypes.GetSources(v1)
	if err != nil {
		log.Println(err)
	}
	domains = configTypes.GetActiveSourceByType(sources, siteGraphType)

	for k := range domains {
		log.Printf("Processing sitegraph file (this can be slow with little feedback): %s", domains[k].URL)
		log.Printf("Downloading sitegraph file: %s", domains[k].URL)

		d, err := getJSON(domains[k].URL)
		if err != nil {
			fmt.Println("error with reading graph JSON: " + domains[k].URL)
		}

		// TODO, how do we quickly validate the JSON-LD files to make sure it is at least formatted well

		sha := common.GetSHA(d) // Don't normalize big files..

		// Upload the file
		log.Printf("Sitegraph file downloaded. Uploading to %s: %s", bucketName, domains[k].URL)

		objectName := fmt.Sprintf("summoned/%s/%s.jsonld", domains[k].Name, sha)
		_, err = graph.LoadToMinio(d, bucketName, objectName, mc)
		if err != nil {
			return objectName, err
		}
		log.Printf("Sitegraph file uploaded to %s. Uploaded : %s", bucketName, domains[k].URL)
		// mill the json-ld to nq and upload to minio
		// we bypass graph.GraphNG which does a time consuming blank node fix which is not required
		// when dealing with a single large file.
		// log.Print("Milling graph")
		//graph.GraphNG(mc, fmt.Sprintf("summoned/%s/", domains[k].Name), v1)
		proc, options := common.JLDProc(v1) // Make a common proc and options to share with the upcoming go funcs
		rdf, err := common.JLD2nq(d, proc, options)
		if err != nil {
			return "", err
		}

		log.Printf("Processed Sitegraph being uploaded to %s: %s", bucketName, domains[k].URL)
		milledName := fmt.Sprintf("milled/%s/%s.rdf", domains[k].Name, sha)
		_, err = graph.LoadToMinio(rdf, bucketName, milledName, mc)
		if err != nil {
			return objectName, err
		}
		log.Printf("Processed Sitegraph Upload to %s complete: %s", bucketName, domains[k].URL)

		// build prov
		// log.Print("Building prov")
		err = StoreProvNG(v1, mc, domains[k].Name, sha, domains[k].URL, "summoned")
		if err != nil {
			return objectName, err
		}

		log.Printf("Loaded: %d", len(d))
	}

	return "Sitegraph(s) processed", err
}

func getJSON(urlloc string) (string, error) {

	urlloc = strings.TrimSpace(urlloc)
	//resp, err := http.Get(url)
	//if err != nil {
	//	return "", fmt.Errorf("GET error: %v", err)
	//}
	/*  https://oih.aquadocs.org/aquadocs.json  fialing with a 403.
	// this is to http 1.1 spec: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Host
	*/

	var client http.Client // why do I make this here..  can I use 1 client?  move up in the loop
	req, err := http.NewRequest("GET", urlloc, nil)
	if err != nil {
		log.Println(err)
	}
	req.Header.Set("User-Agent", "EarthCube_DataBot/1.0")
	u, err := url.Parse(urlloc)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Host", u.Hostname())
	resp, err := client.Do(req)
	if err != nil {
		log.Printf(" error on %s : %s  ", urlloc, err) // print an message containing the index (won't keep order)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status error: %v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %v", err)
	}

	return string(data), nil
}
