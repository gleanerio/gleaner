package graph

import (
	"bytes"
	"fmt"
	"github.com/gleanerio/gleaner/nabu/pkg/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
)

// Drop removes a graph
func Drop(v1 *viper.Viper, g string) ([]byte, error) {
	//spql := v1.GetStringMapString("sparql")
	//spql, _ := config.GetSparqlConfig(v1)
	// d := fmt.Sprintf("DELETE { GRAPH <%s> {?s ?p ?o} } WHERE {GRAPH <%s> {?s ?p ?o}}", g, g)
	d := fmt.Sprintf("DROP GRAPH <%s> ", g)
	pab := []byte(d)

	ep := v1.GetString("flags.endpoint")
	spql, err := config.GetEndpoint(v1, ep, "update")
	if err != nil {
		log.Error(err)
	}

	//req, err := http.NewRequest("POST", spql["endpoint"], bytes.NewBuffer(pab))
	req, err := http.NewRequest(spql.Method, spql.URL, bytes.NewBuffer(pab))
	if err != nil {
		log.Error(err)
	}
	req.Header.Set("Content-Type", "application/sparql-update")
	// req.Header.Set("Content-Type", "application/sparql-results+xml")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("response Body:", string(body))
		log.Error("response Status:", resp.Status)
		log.Error("response Headers:", resp.Header)
	}

	log.Trace(string(body))

	return body, err
}
