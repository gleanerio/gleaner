package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Summoner struct {
	After    string
	Mode     string // full || diff:  If diff compare what we have currently in gleaner to sitemap, get only new, delete missing
	Threads  int
	Delay    int    // milliseconds (1000 = 1 second) to delay between calls (will FORCE threads to 1)
	Headless string // URL for headless see docs/headless
}

var SummonerTemplate = map[string]interface{}{
	"summoner": map[string]string{
		"after":    "",     // "21 May 20 10:00 UTC"
		"mode":     "full", //full || diff:  If diff compare what we have currently in gleaner to sitemap, get only new, delete missing
		"threads":  "5",
		"delay":    "10000",                 // milliseconds (1000 = 1 second) to delay between calls (will FORCE threads to 1)
		"headless": "http://127.0.0.1:9222", // URL for headless see docs/headless
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
	return summoner, err
}
