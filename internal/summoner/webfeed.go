package summoner

import (
	log "github.com/sirupsen/logrus"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/spf13/viper"
)

type RunPROV struct {
	Domain          string // main domain of site indexed
	URL             string // sitemap URL
	Name            string // Short reference name of org indexed
	PID             string // PID of org indexed
	ProperName      string // proper name of org indexed
	AgentName       string // short name of indexing agent
	AgentPID        string // PID of indexing agent
	AgentProperName string // proper name of agent indexing
	EndTime         time.Time
	// ResURL   []string //
	// S3SELECTURLDESCRIPTION []string   or are these a map[string]string
}

func RunFeed(v1 *viper.Viper, mc *minio.Client, et time.Time, ru map[string][]string, hru map[string][]string) error {

	// blend the two maps together

	// for each "org" in the map, build a JSON-LD based DataFeed

	// I could get the description by checking the prov and looking for the URL vis s3select
	// then looking at that object in teh store and pulling description.
	// Need
	// 1) s3select for object ID associated with URL from sitemap
	// 2) s3select for description of a given objectID

	log.Println(len(t))

	return nil
}

const t = `{
    "@context": "https://schema.org/",
    "@type": "WebSite",
    "url": "{{.Domain}}",
    "creator": {
        "@id": "https://gleaner.io/id/org/{{.Name}}",
        "@type": "Organization",
		"url": "{{.Domain}}",
		"name": "{{.Name}}",
		 "identifier": {
			"@type": "PropertyValue",
			"@id": "{{.PID}}",
			"propertyID": "https://registry.identifiers.org/registry/doi",
			"url": "{{.PID}}",
			"description": "{{.ProperName}}"
		}
    },
    "potentialAction": {
        "@type": "ConsumeAction",
        "actionStatus": "CompletedActionStatus",
        "target": {
            "@type": "EntryPoint",
            "urlTemplate": "{{.URL}}",
            "contentType": "application/xml"
        },
        "agent": {
            "@id": "{{.AgentPID}}",
            "@type": "Organization",
			"name": {{.AgentName}}
            "identifier": {
                "@type": "PropertyValue",
                "propertyID": "https://registry.identifiers.org/registry/doi",
                "url": "{{.AgentPID}}",
                "description": "{{.AgentProperName}}"
            },
            "endTime": "{{.EndTime}}",
            "result": {
                "@type": "DataFeed",
                "name": "Gleaner DataFeed for {{.ProperName}}",
                "dateModified": "{{.EndTime}}",
                "dataFeedElement": [
                    {
                        "@type": "DataFeedItem",
                        "dateCreated": "{{.EndTime}}",
                        "item": {
                            "@type": "WebPage",
                            "url": "{{.ResURL}}",
                            "description": "{{.S3SELECTURLDESCRIPTION}}"
                        }
                    }
                ]
            }
        }
    }
}`
