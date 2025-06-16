package graph

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gleanerio/gleaner/nabu/pkg/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// MakeURN formats a URN following the ADR 0001-URN-decision.md  which at the
// time of this coding resulted in     urn:{program}:{organization}:{provider}:{sha}
func MakeURN(v1 *viper.Viper, s string) (string, error) {
	gcfg, _ := config.GetImplNetworkConfig(v1)

	var (
		g   string // build the URN for the graph context string we use
		err error
	)
	s3c := ""
	check := prefixTransform(s) // change "summoned" to "data" if summoned is in the object prefix
	if strings.Contains(check, "orgs/") {
		s3c = check
	} else {
		s3c = getLastThree(check)
	}

	g = fmt.Sprintf("urn:gleaner.io:%s:%s", gcfg.Orgname, s3c) // form the URN

	//fmt.Printf("=MakeURN===========> %s \n\n", g)

	return g, err
}

// MakeURNPrefix formats a URN following the ADR 0001-URN-decision.md  which at the
// time of this coding resulted in   urn:{engine}:{implnet}:{source}:{type}:{sha}
// the "prefix" version only returns the prefix part of the urn, for use in the prune
// command
func MakeURNPrefix(v1 *viper.Viper, prefix string) (string, error) {
	gcfg, _ := config.GetImplNetworkConfig(v1)

	var (
		g   string // build the URN for the graph context string we use
		err error
	)

	if prefix == "orgs" {
		g = fmt.Sprintf("urn:gleaner.io:%s:orgs", gcfg.Orgname)

	} else {
		check := prefixTransform(prefix)
		ps := strings.Split(check, "/")
		g = fmt.Sprintf("urn:gleaner.io:%s:%s:%s", gcfg.Orgname, ps[len(ps)-1], ps[len(ps)-2])
	}

	//fmt.Printf("=Prefix===========> %s \n\n", g)

	return g, err
}

// prefixTransform  In this code, the prefix will be coming in with something like
// summoned or prov.  In our 0001-URN-decision.md document, we want the urn to be like
// urn:gleaner.io:oih:edmo:prov:0255293683036aac2a95a2479cc841189c0ac3f8
// or
// urn:gleaner.io:iow:counties0:data:00010f9f071c39fcc0ca73eccad7470b675cd8a3
// this means that the string "summoned" needs to be mapped to "data".  However,
// we use prov for both the path in the S3 and the URN structure.  So in this
// location we need to convert summoned to prov
func prefixTransform(str string) string {
	if !strings.Contains(str, "summoned/") {
		return str
	}

	return strings.Replace(str, "summoned/", "data/", -1)
}

// getLastThree
// split the string and take last two segments, but flip to match URN for ADR 0001-URN-decision.md
func getLastThree(s string) string {
	extension := filepath.Ext(s) // remove the extension regardless of what it is
	s = strings.TrimSuffix(s, extension)

	sr := strings.Replace(s, "/", ":", -1) // replace / with :
	parts := strings.Split(sr, ":")        // Split the string on the ":" character.

	lastThree := parts[len(parts)-3:] // Get the last three elements.

	//flip the last two elements
	index1 := 0
	index2 := 1

	// Ensure indices are within the array bounds
	if index1 >= 0 && index1 < len(lastThree) && index2 >= 0 && index2 < len(lastThree) {
		// Swap the elements
		lastThree[index1], lastThree[index2] = lastThree[index2], lastThree[index1]
	} else {
		log.Println("error in urn formation trying to flip indices on object prefix")
	}

	s2c := strings.Join(lastThree, ":")

	return s2c
}

// getLastTwo from chatGPT

func getLastTwo(s string) string {
	// Split the string on the ":" character.
	parts := strings.Split(s, ":")

	// Get the last two elements.
	lastTwo := parts[len(parts)-2:]

	// Join the last two elements and return the result.
	return strings.Join(lastTwo, ":")
}
