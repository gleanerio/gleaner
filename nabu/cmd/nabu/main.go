package main

import (
	"github.com/gleanerio/gleaner/nabu/pkg/cli"
	"github.com/gleanerio/gleaner/nabu/pkg/nabuinternal/common"
)

func init() {
	common.InitLogging()
}

func main() {
	cli.Execute()
}
