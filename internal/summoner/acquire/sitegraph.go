package acquire

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/boltdb/bolt"
	configTypes "github.com/gleanerio/gleaner/internal/config"

	"github.com/gleanerio/gleaner/internal/common"
	"github.com/gleanerio/gleaner/internal/millers/graph"
	"github.com/minio/minio-go/v7"
	"github.com/spf13/viper"
)

type Sources = configTypes.Sources

const siteGraphType = "sitegraph"

// GetGraph downloads pre-built site graphs
func GetGraph(mc *minio.Client, v1 *viper.Viper, db *bolt.DB) (string, error) {
	// read config file
	var mcfg configTypes.Summoner
	mcfg, err := configTypes.ReadSummmonerConfig(v1.Sub("summoner"))
	if err != nil {
		log.Println(err)
	}

	bucketName, err := configTypes.GetBucketName(v1) //miniocfg["bucket"] //   get the top level bucket for all of gleaner operations from config file
	if err != nil {
		log.Println(err)
	}

	var domains []Sources // get the sitegraph entry from config file

	sources, err := configTypes.ParseSourcesConfig(v1)
	if err != nil {
		log.Println(err)
	}
	domains = configTypes.GetSourceByType(sources, siteGraphType)

	for k := range domains {
		log.Printf("Processing a sitegraph file (this can be slow with little feedback): %s", domains[k].URL)

		// make the bucket (if it doesn't exist)
		db.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucket([]byte(domains[k].Name))
			if err != nil {
				return fmt.Errorf("create bucket: %s", err)
			}
			return nil
		})

		// TODO  issue 44 at this point I need to see if we are diff and remove items have already
		log.Println(domains)

		if mcfg.Mode == "diff" {
			// sitegraphs are a single thing, if we have them we simply want to break out of the for loop

			e := false
			db.View(func(tx *bolt.Tx) error {
				b := tx.Bucket([]byte(domains[k].Name))
				value := b.Get([]byte(domains[k].URL))
				if value != nil {
					fmt.Println("key exists and in diff mode we will not proceed")
					e = true
				}
				return err
			})

			if e {
				break
			} else {
				log.Println("We don't have a record of this sitegraph being downloaded, downloading now")
			}

		}
		// end of check

		d, err := getJSON(domains[k].URL)
		if err != nil {
			fmt.Println("error with reading graph JSON: " + domains[k].URL)
		}

		// TODO, how do we quickly validate the JSON-LD files to make sure it is at least formatted well

		sha := common.GetSHA(d) // Don't normalize big files..

		// Upload the file
		log.Printf("Uploading sitegraph file: %s :: %s", domains[k].Name, domains[k].URL)
		objectName := fmt.Sprintf("summoned/%s/%s.jsonld", domains[k].Name, sha)
		_, err = graph.LoadToMinio(d, bucketName, objectName, mc)
		if err != nil {
			return objectName, err
		}

		// Mill the json-ld to nq and upload to minio  (this can be slow... should we let Nabu worry about sitegraphs rather than hold up indexing?)
		// We bypass graph.GraphNG which does a time consuming blank node fix which is not required  when dealing with a single large file.
		log.Printf("Milling sitegraph file: %s :: %s", domains[k].Name, domains[k].URL)
		proc, options := common.JLDProc(v1) // Make a common proc and options to share with the upcoming go funcs
		rdf, err := common.JLD2nq(d, proc, options)
		if err != nil {
			return "", err
		}

		milledName := fmt.Sprintf("milled/%s/%s.rdf", domains[k].Name, sha)
		_, err = graph.LoadToMinio(rdf, bucketName, milledName, mc)
		if err != nil {
			return objectName, err
		}

		// build prov
		log.Printf("Building sitegraph prov: %s :: %s", domains[k].Name, domains[k].URL)
		err = StoreProvNG(v1, mc, domains[k].Name, sha, domains[k].URL, "summoned")
		if err != nil {
			return objectName, err
		}

		// issue 44  Add the URL if we did pull it down
		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(domains[k].Name))
			log.Printf("%s    %s", domains[k].URL, sha)
			err := b.Put([]byte(domains[k].URL), []byte(sha))
			if err != nil {
				log.Println(err)
			}
			return err
		})

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

	var client http.Client
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
