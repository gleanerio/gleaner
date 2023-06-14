package summoner

import (
	"fmt"
	"github.com/gleanerio/gleaner/internal/common"
	"github.com/minio/minio-go/v7"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gleanerio/gleaner/internal/summoner/acquire"
	"github.com/spf13/viper"
)

func RunStatsOutput(runStats *common.RunStats) {
	fmt.Print(runStats.Output())
	const layout = "2006-01-02-15-04-05"
	t := time.Now()
	lf := fmt.Sprintf("%s/gleaner-runstats-%s.log", common.Logpath, t.Format(layout))

	LogFile := lf // log to custom file
	logFile, err := os.OpenFile(LogFile, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	logFile.WriteString(runStats.Output())
	logFile.Close()
}

// Summoner pulls the resources from the data facilities
// func Summoner(mc *minio.Client, cs utils.Config) {
func Summoner(mc *minio.Client, v1 *viper.Viper) {

	st := time.Now()
	log.Info("Summoner start time:", st) // Log the time at start for the record
	runStats := common.NewRunStats()

	// Retrieve API urls
	apiSources, err := acquire.RetrieveAPIEndpoints(v1)
	if err != nil {
		log.Error("Error getting API endpoint sources:", err)
	} else if len(apiSources) > 0 {
		acquire.RetrieveAPIData(apiSources, mc, runStats, v1)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		runStats.StopReason = "User Interrupt or Fatal Error"
		RunStatsOutput(runStats)
		os.Exit(1)
	}()

	// Get a list of resource URLs that do and don't require headless processing
	ru, err := acquire.ResourceURLs(v1, mc, false)
	if err != nil {
		log.Info("Error getting urls that do not require headless processing:", err)
	}
	// just report the error, and then run gathered urls
	if len(ru) > 0 {
		acquire.ResRetrieve(v1, mc, ru, runStats) // TODO  These can be go funcs that run all at the same time..
	}

	hru, err := acquire.ResourceURLs(v1, mc, true)
	if err != nil {
		log.Info("Error getting urls that require headless processing:", err)
	}
	// just report the error, and then run gathered urls
	if len(hru) > 0 {
		log.Info("running headless:")
		acquire.HeadlessNG(v1, mc, hru, runStats)
	}

	// Time report
	et := time.Now()
	diff := et.Sub(st)
	log.Info("Summoner end time:", et)
	log.Info("Summoner run time:", diff.Minutes())
	runStats.StopReason = "Complete"
	RunStatsOutput(runStats)
	// What do I need to the "run" prov
	// the URLs indexed  []string
	// the graph generated?  "version" the graph by the build date
	// pass ru, hru, and v1 to a run prov function.
	//	RunFeed(v1, mc, et, ru, hru)  // DEV:   hook for building feed  (best place for it?)

}
