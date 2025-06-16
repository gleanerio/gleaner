package config

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// frig frig... do not use lowercase... those are private variables
type ImplNetwork struct {
	Orgname string
}

var implNetworkTemplate = map[string]interface{}{
	"implementation_network": map[string]string{
		"orgname": "indexer",
	},
}

func GetImplNetworkConfig(viperConfig *viper.Viper) (ImplNetwork, error) {
	sub := viperConfig.Sub("implementation_network")
	return readImpleNetworkConfig(sub)
}

// use config.Sub("gleaner)
func readImpleNetworkConfig(implementation_networkSubtress *viper.Viper) (ImplNetwork, error) {
	var gleanerCfg ImplNetwork
	if implementation_networkSubtress == nil {
		// trace, otherwise  goes off for every item
		log.Trace("No Implementation Network in config file: Add \n implementation_network: \n   orgname: NAME    ")
		implementation_networkSubtress = viper.New()
		//for key, value := range implNetworkTemplate {
		//	implementation_networkSubtress.Set(key, value)
		//}
	}
	for key, value := range implNetworkTemplate {
		implementation_networkSubtress.SetDefault(key, value)
	}
	implementation_networkSubtress.BindEnv("orgname", "IMPLEMENTATION_NETWORK_ORGNAME")

	implementation_networkSubtress.AutomaticEnv()
	// config already read. substree passed
	err := implementation_networkSubtress.Unmarshal(&gleanerCfg)
	if err != nil {
		panic(fmt.Errorf("error when parsing gleaner config: %v", err))
	}
	return gleanerCfg, err
}
