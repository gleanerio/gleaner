package acquire

import (
	log "github.com/sirupsen/logrus"
	"strconv"

	"github.com/spf13/viper"
)

func Threadcount(v1 *viper.Viper) (int64, error) {

	mcfg := v1.GetStringMapString("summoner")
	tc, err := strconv.ParseInt(mcfg["threads"], 10, 64)
	if err != nil {
		log.Error(err)
		log.Error("Could not convert threads from config file to an int. setting to 5")
		tc = 5
	}

	return tc, err

}

func Delayrequest(v1 *viper.Viper) (int64, error) {
	mcfg := v1.GetStringMapString("summoner")

	var err error

	delay := mcfg["delay"]
	var dt int64
	if delay != "" {
		log.Debug("Delay set to: ", delay, "milliseconds")
		dt, err = strconv.ParseInt(delay, 10, 64)
		if err != nil {
			//log.Panic(err, "Could not convert delay from config file to a value")
			log.Error(err, "Could not convert delay from config file to a value. setting to zero")
			dt = 0
		}
	} else {
		dt = 0
	}

	return dt, err
}
