package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Sparql struct {
	Endpoint       string
	EndpointBulk   string
	EndpointMethod string
	ContentType    string
	Authenticate   bool
	Username       string
	Password       string
}

var sparqlTemplate = map[string]interface{}{
	"sparql": map[string]string{
		"endpoint":       "http://example.org:3030/testing/sparql",
		"endpointBulk":   "http://example.org:3030/testing/data",
		"endpointMethod": "POST",
		"contentType":    "application/n-quads",
		"authenticate":   "False",
		"username":       "",
		"password":       "",
	},
}

func DEPRECATEDGetSparqlConfig(viperConfig *viper.Viper) (Sparql, error) {
	sub := viperConfig.Sub("sparql")
	return readSparqlConfig(sub)
}

func readSparqlConfig(viperSubtree *viper.Viper) (Sparql, error) {
	var sparql Sparql
	for key, value := range sparqlTemplate {
		viperSubtree.SetDefault(key, value)
	}
	viperSubtree.BindEnv("endpoint", "SPARQL_ENDPOINT")
	viperSubtree.BindEnv("endpointBulk", "SPARQL_ENDPOINTBULK")
	viperSubtree.BindEnv("endpointMethod", "SPARQL_ENDPOINTMETHOD")
	viperSubtree.BindEnv("contentType", "SPARQL_CONTENTTYPE")
	viperSubtree.BindEnv("authenticate", "SPARQL_AUTHENTICATE")
	viperSubtree.BindEnv("username", "SPARQL_USERNAME")
	viperSubtree.BindEnv("password", "SPARQL_PASSWORD")

	viperSubtree.AutomaticEnv()
	// config already read. substree passed
	err := viperSubtree.Unmarshal(&sparql)
	if err != nil {
		panic(fmt.Errorf("error when parsing sparql endpoint config: %v", err))
	}
	return sparql, err
}
