package acquire

import (
	"bytes"
	"context"
	"fmt"
	configTypes "github.com/gleanerio/gleaner/internal/config"
	log "github.com/sirupsen/logrus"
	"text/template"
	"time"

	"github.com/gleanerio/gleaner/internal/common"
	"github.com/gleanerio/gleaner/internal/objects"
	"github.com/minio/minio-go/v7"
	"github.com/spf13/viper"
)

// ProvData is the struct holding the prov data for a summoned data graph
type ProvData struct {
	RESID  string
	SHA256 string
	PID    string
	SOURCE string
	DATE   string
	RUNID  string
	URN    string
	PNAME  string
	DOMAIN string
}

// Gleaner prov data needs to align with the ADR published in Nabu for named graphs patterns.
// This can be found at (https://github.com/gleanerio/nabu/blob/df-dev/decisions/0001-URN-decision.md)
// which is a reference that needs updating when this is pushed to master.
// The pattern should be like:
// urn:{program}:{organization}:{provider}:{sha}
// urn:gleaner.io:earthcube:edmo:0255293683036aac2a95a2479cc841189c0ac3f8

func StoreProvNG(v1 *viper.Viper, mc *minio.Client, k, sha, urlloc, objprefix string) error {
	bucketName, err := configTypes.GetBucketName(v1)

	p, err := provOGraph(v1, k, sha, urlloc, objprefix)
	if err != nil {
		return err
	}

	// Moved to the normalized sha value since normalized sha only valid for JSON-LD
	provsha := common.GetSHA(p)
	// provsha, err := common.GetNormSHA(p, v1) // Moved to the normalized sha value
	// if err != nil {
	// 	log.Printf("ERROR: URL: %s Action: Getting normalized sha  Error: %s\n", urlloc, err)
	// 	return err
	// }

	b := bytes.NewBufferString(p)

	objectName := fmt.Sprintf("prov/%s/%s.jsonld", k, provsha) // k is the name of the provider from config
	usermeta := make(map[string]string)                        // what do I want to know?
	usermeta["url"] = urlloc
	usermeta["sha1"] = sha // recall this is the sha of the about object, not the prov graph itself

	contentType := JSONContentType

	// Upload the file with FPutObject
	_, err = mc.PutObject(context.Background(), bucketName, objectName, b, int64(b.Len()), minio.PutObjectOptions{ContentType: contentType, UserMetadata: usermeta})
	if err != nil {
		log.Errorf("%s: %s", objectName, err)
	}

	return err
}

// provOGraph  Against better judgment rather than build triples, I'll just template build them
func provOGraph(v1 *viper.Viper, k, sha, urlloc, objprefix string) (string, error) {
	// read config file
	miniocfg := v1.GetStringMapString("minio")
	bucketName := miniocfg["bucket"] //   get the top level bucket for all of gleaner operations from config file

	// get the time
	currentTime := time.Now() // date := currentTime.Format("2006-01-02")

	// open the config to get the runID later
	mcfg := v1.GetStringMapString("gleaner")
	domains := objects.SourcesAndGraphs(v1)

	pid := "unknown"
	pname := "unknown"
	domain := "unknown"
	for i := range domains {
		if domains[i].Name == k {
			pid = domains[i].PID
			pname = domains[i].ProperName
			domain = domains[i].Domain
		}
	}

	// TODO make an extracted function to share with nabu
	// make the URN string

	// urn:{program}:{organization}:{provider}:{sha}
	//
	//urn:gleaner.io:earthcube:edmo:0255293683036aac2a95a2479cc841189c0ac3f8
	var objectURN string
	objectURN = fmt.Sprintf("%s:%s", k, sha)

	//if strings.Contains(objprefix, "summoned") {
	//	objectURN = fmt.Sprintf("summoned:%s:%s", k, sha)
	//} else if strings.Contains(objprefix, "milled") {
	//	objectURN = fmt.Sprintf("milled:%s:%s", k, sha)
	//} else {
	//	return "", errors.New("no valid prov object prefix")
	//}

	// build the struct to pass to the template parser
	gp := fmt.Sprintf("urn:%s:%s", bucketName, objectURN)
	td := ProvData{RESID: urlloc, SHA256: sha, PID: pid, SOURCE: k,
		DATE:   currentTime.Format("2006-01-02"),
		RUNID:  mcfg["runid"],
		URN:    gp,
		PNAME:  pname,
		DOMAIN: domain}

	var doc bytes.Buffer
	t, err := template.New("prov").Parse(provTemplate())
	if err != nil {
		log.Error("Prov Failure: Cannot parse or read template")
		//panic(err) // don't stop processing for a bad prov
		return "", err
	}
	err = t.Execute(&doc, td)
	if err != nil {
		//panic(err) // don't stop processing for a bad prov
		log.Error("Prov Failure")
		return "", err
	}

	return doc.String(), err
}

func provTemplate() string {

	t := `{
		"@context": {
		  "rdf": "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
		  "prov": "http://www.w3.org/ns/prov#",
		  "rdfs": "http://www.w3.org/2000/01/rdf-schema#"
		},
		"@graph": [
		  {
			"@id": "{{.PID}}",
			"@type": "prov:Organization",
			"rdf:name": "{{.PNAME}}",
			"rdfs:seeAlso": "{{.DOMAIN}}"
		  },
		  {
			"@id": "{{.RESID}}",
			"@type": "prov:Entity",
			"prov:wasAttributedTo": {
			  "@id": "{{.PID}}"
			},
			"prov:value": "{{.RESID}}"
		  },
		  {
			"@id": "https://gleaner.io/id/collection/{{.SHA256}}",
			"@type": "prov:Collection",
			"prov:hadMember": {
			  "@id": "{{.RESID}}"
			}
		  },
		  {
			"@id": "{{.URN}}",
			"@type": "prov:Entity",
			"prov:value": "{{.SHA256}}.jsonld"
		  },
		  {
			"@id": "https://gleaner.io/id/run/{{.SHA256}}",
			"@type": "prov:Activity",
			"prov:endedAtTime": {
			  "@value": "{{.DATE}}",
			  "@type": "http://www.w3.org/2001/XMLSchema#dateTime"
			},
			"prov:generated": {
			  "@id": "{{.URN}}"
			},
			"prov:used": {
			  "@id": "https://gleaner.io/id/collection/{{.SHA256}}"
			}
		  }
		]
	  }`

	return t

}
