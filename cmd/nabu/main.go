package main

import (
	"github.com/gleanerio/gleaner/nabu/pkg/nabuinternal/common"
	"github.com/gleanerio/gleaner/pkg/cli"
)

func init() {
	common.InitLogging()
}

func main() {
	cli.Execute()
}
