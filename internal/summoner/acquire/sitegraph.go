package acquire

import (
	"fmt"
	configTypes "github.com/earthcubearchitecture-project418/gleaner/internal/config"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gleanerio/gleaner/internal/common"
	"github.com/gleanerio/gleaner/internal/millers/graph"
	"github.com/gleanerio/gleaner/internal/objects"
	"github.com/minio/minio-go/v7"
	"github.com/spf13/viper"
)

type Sources = configTypes.Sources

const siteGraphType = "sitegraph"

// GetGraph downloads pre-built site graphs
func GetGraph(mc *minio.Client, v1 *viper.Viper) (string, error) {
	// read config file
	miniocfg := v1.GetStringMapString("minio")
	bucketName := miniocfg["bucket"] //   get the top level bucket for all of gleaner operations from config file

	// get the sitegraph entry from config file
	var domains []Sources
	//err := v1.UnmarshalKey("sitegraphs", &domains)

	sources, err := configTypes.ParseSourcesConfig(v1)
	if err != nil {
		log.Println(err)
	}
	domains = configTypes.GetSourceByType(sources, siteGraphType)

	for k := range domains {
		log.Printf("Processing sitegraph file (this can be slow with little feedback): %s", domains[k].URL)

		d, err := getJSON(domains[k].URL)
		if err != nil {
			fmt.Println("error with reading graph JSON")
		}

		// TODO, how do we quickly validate the JSON-LD files to make sure it is at least formatted well

		sha := common.GetSHA(d) // Don't normalize big files..

		// Upload the file
		// log.Print("Uploading file")
		objectName := fmt.Sprintf("summoned/%s/%s.jsonld", domains[k].Name, sha)
		_, err = graph.LoadToMinio(d, bucketName, objectName, mc)
		if err != nil {
			return objectName, err
		}

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

		milledName := fmt.Sprintf("milled/%s/%s.rdf", domains[k].Name, sha)
		_, err = graph.LoadToMinio(rdf, bucketName, milledName, mc)
		if err != nil {
			return objectName, err
		}

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

func getJSON(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("GET error: %v", err)
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
