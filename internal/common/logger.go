package common

import (
	"errors"
	"fmt"
	"github.com/orandin/lumberjackrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io"
	"os"
	"time"
)

func InitLogging() {
	logpath := "logs"
	if _, err := os.Stat(logpath); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(logpath, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
	// name the file with the date and time
	const layout = "2006-01-02-15-04-05"
	t := time.Now()
	lf := fmt.Sprintf("%s/gleaner-%s.log", logpath, t.Format(layout))

	LogFile := lf // log to custom file
	logFile, err := os.OpenFile(LogFile, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
		return
	}

	log.SetFormatter(&log.JSONFormatter{}) // Log as JSON instead of the default ASCII formatter.
	log.SetReportCaller(true)              // include file name and line number
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
	//log.SetOutput(logFile)

	// second file for issues

}
func SetLogLevel(v1 *viper.Viper) {
	log.SetLevel(log.InfoLevel)
}

// this is a fixed file format for outputing to a file.
// Error... Generally what we want
// Info... Level we set. This will give us the Starting and Ending and other info
// Debug --- This will give us all the ins and outs of the summoning
// Trace --- all the details
func LogIssues(v1 *viper.Viper, source string) (*log.Logger, error) {
	logpath := "logs"
	if _, err := os.Stat(logpath); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(logpath, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
	// name the file with the date and time
	const layout = "2006-01-02-15-04-05"
	t := time.Now()

	logger := log.New()

	issuefile := fmt.Sprintf("%s/repo-%s-issues-%s.log", logpath, source, t.Format(layout))
	allfile := fmt.Sprintf("%s/repo-%s-loaded-%s.log", logpath, source, t.Format(layout))

	logger.SetFormatter(&log.TextFormatter{DisableTimestamp: true}) // Log as JSON instead of the default ASCII formatter.
	logger.SetReportCaller(false)                                   // disable include file name and line number
	logFile, err := os.OpenFile(os.DevNull, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	logger.SetOutput(logFile)
	logger.SetLevel(log.TraceLevel) // this effects the lumberjacks

	// second file for issues

	hook, err := lumberjackrus.NewHook(
		&lumberjackrus.LogFile{
			Filename:   allfile,
			MaxSize:    100,
			MaxBackups: 1,
			MaxAge:     1,
			Compress:   false,
			LocalTime:  false,
		},
		log.DebugLevel, //log.TraceLevel,  // needs to be configurable
		&log.TextFormatter{
			DisableColors:    true,
			FullTimestamp:    false,
			DisableTimestamp: true,
			//FieldMap: &log.FieldMap{
			//	"level": "@level",
			//	"msg":   "@message",
			//}
		},
		&lumberjackrus.LogFileOpts{
			//log.TraceLevel: &lumberjackrus.LogFile{
			//	Filename: allfile,
			//},
			log.ErrorLevel: &lumberjackrus.LogFile{
				Filename:   issuefile,
				MaxSize:    100,   // optional
				MaxBackups: 1,     // optional
				MaxAge:     1,     // optional
				Compress:   false, // optional
				LocalTime:  false, // optional
			},
		},
	)

	if err != nil {
		panic(err)
	}

	logger.AddHook(hook)

	return logger, nil
}
