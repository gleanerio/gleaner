package main

import (
	"fmt"
	"github.com/gleanerio/gleaner/pkg/cli"
)

var VERSION string

func main() {
	fmt.Println("version: ", VERSION)
	cli.Execute()
}
