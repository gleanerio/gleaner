package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type EndPoint struct {
	Service      string
	Baseurl      string
	Type         string
	Authenticate bool
	Username     string
	Password     string
	Modes        []Mode
}

type Mode struct {
	Action string
	Suffix string
	Accept string
	Method string
}

type ServiceMode struct {
	Service      string
	URL          string // combined Baseurl + Suffix
	Type         string
	Authenticate bool
	Username     string
	Password     string
	Accept       string
	Method       string
}

func GetEndPointsConfig(v1 *viper.Viper) ([]EndPoint, error) {
	var subtreeKey = "endpoints"
	var endpointsCfg []EndPoint

	if v1 == nil {
		return nil, fmt.Errorf("GetEndPointsConfig: viperConfig is nil")
	}

	err := v1.UnmarshalKey(subtreeKey, &endpointsCfg)
	if err != nil {
		log.Fatal("error when parsing ", subtreeKey, " config: ", err)
		//No sources, so nothing to run
	}

	return endpointsCfg, err
}

func GetEndpoint(v1 *viper.Viper, set, servertype string) (ServiceMode, error) {
	// TODO change the return to be this
	sm := ServiceMode{}
	var err error

	epcfg, err := GetEndPointsConfig(v1)
	if err != nil {
		log.Fatalf("error getting endpoint node in config %v", err)
	}

	// TODO if set nil
	if set == "" && len(epcfg) != 1 {
		// this is an error, they need to specify a service
	}
	if set == "" && len(epcfg) == 1 {
		set = epcfg[0].Service
	}

	// loop through our endpointsfor the set and then loop through for the mode we want
	for _, item := range epcfg {
		if item.Service == set {
			for _, m := range item.Modes {
				if m.Action == servertype {
					// Now,collect the set and mode into a new
					// ServiceMode struct so the approach is still spql.PROPERTY in the code
					sm.Service = item.Service
					sm.URL = item.Baseurl + m.Suffix
					sm.Type = item.Type
					sm.Authenticate = item.Authenticate
					sm.Username = item.Username
					sm.Password = item.Password
					sm.Accept = m.Accept
					sm.Method = m.Method
					return sm, nil // return the item if found
				}
			}
		}
	}

	// If at this point we don't have a SPARQL endpoint, then we might as well stop, there
	// is not much Nabu can do
	// TODO could also check that the URL is a valid http structure
	if sm.URL == "" {
		log.Fatalf("FATAL: error getting SPARQL endpoint node from config")
	}

	return sm, err
}
