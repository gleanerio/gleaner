package main

import (
	"github.com/gleanerio/gleaner/pkg/cli"
	log "github.com/sirupsen/logrus"
)

var VERSION string

func main() {
	log.Print("version: ", VERSION)
	cli.Execute()
}
