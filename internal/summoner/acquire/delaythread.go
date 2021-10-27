package acquire

import (
	"log"
	"strconv"

	"github.com/spf13/viper"
)

func Threadcount(v1 *viper.Viper) (int64, error) {

	mcfg := v1.GetStringMapString("summoner")
	tc, err := strconv.ParseInt(mcfg["threads"], 10, 64)
	if err != nil {
		log.Println(err)
		log.Panic("Could not convert threads from config file to an int")
	}

	return tc, err

}

func Delayrequest(v1 *viper.Viper) (int64, error) {
	mcfg := v1.GetStringMapString("summoner")

	var err error

	delay := mcfg["delay"]
	var dt int64
	if delay != "" {
		//log.Printf("Delay set to: %s milliseconds", delay)
		dt, err = strconv.ParseInt(delay, 10, 64)
		if err != nil {
			log.Println(err)
			log.Panic("Could not convert delay from config file to a value")
		}
	} else {
		dt = 0
	}

	return dt, err
}
