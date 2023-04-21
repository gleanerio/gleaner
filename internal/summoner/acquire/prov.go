package acquire

import (
	"bytes"
	"context"
	"fmt"
	"text/template"
	"time"

	configTypes "github.com/gleanerio/gleaner/internal/config"
	log "github.com/sirupsen/logrus"

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

func StoreProvNG(v1 *viper.Viper, mc *minio.Client, k, sha, urlloc, objprefix string) error {
	// read config file
	bucketName, err := configTypes.GetBucketName(v1)

	p, err := provOGraph(v1, k, sha, urlloc, objprefix)
	if err != nil {
		return err
	}

	// Moved to the simple sha value since normalized sha only valid for JSON-LD
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
	usermeta["sha1"] = sha // recall this is the sha of the data graph the prov is about, not the prov graph itself

	contentType := "application/ld+json"

	// Upload the file with FPutObject
	_, err = mc.PutObject(context.Background(), bucketName, objectName, b, int64(b.Len()), minio.PutObjectOptions{ContentType: contentType, UserMetadata: usermeta})
	if err != nil {
		log.Errorf("%s: %s", objectName, err)
		// Fatal?   seriously?    I guess this is the object write, so the run is likely a bust at this point, but this seems a bit much still.
	}

	return err
}

// provOGraph is a simpler provo prov function
// I'll just build from a template for now, but using a real RDF lib to build these triples would be better
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

	// TODO:  There is danger here if this and the URN for the graph from Nabu do not match.
	// We need to modify this to help prevent that from happening.
	// Shouuld align with:  https://github.com/gleanerio/nabu/blob/dev/decisions/0001-URN-decision.md
	gp := fmt.Sprintf("urn:%s:%s:%s", bucketName, k, sha)

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
		return "", err
	}
	err = t.Execute(&doc, td)
	if err != nil {
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
