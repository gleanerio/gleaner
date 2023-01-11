package config

import (
	"fmt"
	"github.com/spf13/viper"
	"strings"
)

type Summoner struct {
	After          string
	Mode           string // full || diff:  If diff compare what we have currently in gleaner to sitemap, get only new, delete missing
	Threads        int
	Delay          int64  // milliseconds (1000 = 1 second) to delay between calls (will FORCE threads to 1)
	Headless       string // URL for headless see docs/
	IdentifierType string // identifiersha, filesha, identifier
}

var SummonerTemplate = map[string]interface{}{
	"summoner": map[string]string{
		"after":          "",     // "21 May 20 10:00 UTC"
		"mode":           "full", //full || diff:  If diff compare what we have currently in gleaner to sitemap, get only new, delete missing
		"threads":        "5",
		"delay":          "10000",                 // milliseconds (1000 = 1 second) to delay between calls (will FORCE threads to 1)
		"headless":       "http://127.0.0.1:9222", // URL for headless see docs/headless
		"identifiertype": JsonSha,
	},
}

func ReadSummmonerConfig(viperSubtree *viper.Viper) (Summoner, error) {
	var summoner Summoner
	for key, value := range SummonerTemplate {
		viperSubtree.SetDefault(key, value)
	}
	viperSubtree.BindEnv("headless", "GLEANER_HEADLESS_ENDPOINT")
	viperSubtree.BindEnv("threads", "GLEANER_THREADS")
	viperSubtree.BindEnv("mode", "GLEANER_MODE")

	viperSubtree.AutomaticEnv()
	// config already read. substree passed
	err := viperSubtree.Unmarshal(&summoner)
	if err != nil {
		panic(fmt.Errorf("error when parsing sparql endpoint config: %v", err))
	}
	if strings.HasSuffix(summoner.Headless, "/") {
		panic(fmt.Errorf("headless warning should not end with / %v", summoner.Headless))
	}
	return summoner, err
}
