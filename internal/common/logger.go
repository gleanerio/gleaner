package common

import (
	"fmt"
	"github.com/orandin/lumberjackrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io"
	"os"
	"time"
)

func InitLogging() {

	// name the file with the date and time
	const layout = "2006-01-02-15-04-05"
	t := time.Now()
	lf := fmt.Sprintf("gleaner-%s.log", t.Format(layout))

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
	// name the file with the date and time
	const layout = "2006-01-02-15-04-05"
	t := time.Now()
	//lf := fmt.Sprintf("gleaner-%s.log", t.Format(layout))
	logger := log.New()

	issuefile := fmt.Sprintf("repo-%s-issues-%s.log", source, t.Format(layout))
	allfile := fmt.Sprintf("repo-%s-loaded-%s.log", source, t.Format(layout))
	//LogFile := issuefile // log to custom file
	//logFile, err := os.OpenFile(LogFile, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	//if err != nil {
	//	log.Panic(err)
	//	return logger, err // could break things if there is an nil value... so...
	//}

	logger.SetFormatter(&log.TextFormatter{DisableTimestamp: true}) // Log as JSON instead of the default ASCII formatter.
	logger.SetReportCaller(false)                                   // disable include file name and line number
	//mw := io.MultiWriter(os.Stdout, logFile)
	//log.SetOutput(mw)
	//logger.SetOutput(logFile)
	logger.SetLevel(log.TraceLevel) // this effects the lumberjacks
	// second file for issues

	//IssueFile := issuefile // log to custom file
	//issueFile, err := os.OpenFile(IssueFile, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	//if err != nil {
	//	log.Panic(err)
	//	return
	//}
	//
	//imw := io.MultiWriter(os.Stdout, issueFile)
	//log.AddHook(&writer.Hook{ // Send logs with level higher than warning to stderr
	//	Writer: imw,
	//	LogLevels: []log.Level{
	//		log.PanicLevel,
	//		log.FatalLevel,
	//		log.ErrorLevel,
	//		log.WarnLevel,
	//	},
	//})

	hook, err := lumberjackrus.NewHook(
		&lumberjackrus.LogFile{
			Filename:   allfile,
			MaxSize:    100,
			MaxBackups: 1,
			MaxAge:     1,
			Compress:   false,
			LocalTime:  false,
		},
		log.TraceLevel,
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
