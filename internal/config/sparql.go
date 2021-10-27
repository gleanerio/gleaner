package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Sparql struct {
	Endpoint string
}

var sparqlTemplate = map[string]interface{}{
	"sparql": map[string]string{
		"endpoint":     "http://localhost/blazegraph/namespace/nabu/sparql",
		"authenticate": "False",
		"username":     "",
		"password":     "",
	},
}

func ReadSparqlConfig(viperSubtree *viper.Viper) (Sparql, error) {
	var sparql Sparql
	for key, value := range sparqlTemplate {
		viperSubtree.SetDefault(key, value)
	}
	viperSubtree.BindEnv("endpoint", "SPARQL_ENDPOINT")
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
