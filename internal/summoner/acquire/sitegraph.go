package acquire

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/earthcubearchitecture-project418/gleaner/internal/common"
	"github.com/earthcubearchitecture-project418/gleaner/internal/millers/graph"
	"github.com/earthcubearchitecture-project418/gleaner/internal/objects"
	"github.com/minio/minio-go/v7"
	"github.com/spf13/viper"
)

// GetGraph downloads pre-built site graphs
func GetGraph(mc *minio.Client, v1 *viper.Viper) (string, error) {
	// get the sitegraph entry from config file
	var domains []objects.Sources
	err := v1.UnmarshalKey("sitegraphs", &domains)
	if err != nil {
		log.Println(err)
	}

	for k := range domains {
		log.Println(domains[k].URL)

		d, err := getJSON(domains[k].URL)
		if err != nil {
			fmt.Println("error with reading graph JSON")
		}

		// TODO, how do we quickly validate the JSON-LD files to make sure it is at least formatted well

		sha := common.GetSHA(d) // Don't normalize big files..

		// Upload the file
		log.Print("Uploading file")
		objectName := fmt.Sprintf("summoned/%s/%s.jsonld", domains[k].Name, sha)
		_, err = graph.LoadToMinio(d, "gleaner", objectName, mc)
		if err != nil {
			return objectName, err
		}

		// mill the json-ld to nq and upload to minio
		// we bypass graph.GraphNG which does a time consuming blank node fix which is not required
		// when dealing with a single large file.
		log.Print("Milling graph")
		//graph.GraphNG(mc, fmt.Sprintf("summoned/%s/", domains[k].Name), v1)
		proc, options := common.JLDProc(v1) // Make a common proc and options to share with the upcoming go funcs
		rdf, err := common.JLD2nq(d, proc, options)
		if err != nil {
			return "", err
		}

		milledName := fmt.Sprintf("milled/%s/%s.rdf", domains[k].Name, sha)
		_, err = graph.LoadToMinio(rdf, "gleaner", milledName, mc)
		if err != nil {
			return objectName, err
		}

		// build prov
		log.Print("Building prov")
		err = StoreProvNG(v1, mc, domains[k].Name, sha, domains[k].URL, "summoned")
		if err != nil {
			return objectName, err
		}

		log.Println(len(d))
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
