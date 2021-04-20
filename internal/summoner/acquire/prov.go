package acquire

import (
	"bytes"
	"fmt"
	"log"
	"text/template"
	"time"

	"github.com/earthcubearchitecture-project418/gleaner/internal/common"
	"github.com/minio/minio-go"
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

// StoreProv creates and stores a prov record for each JSON-LD data graph
func StoreProv(v1 *viper.Viper, mc *minio.Client, k, sha, urlloc string) error {
	var (
		buf    bytes.Buffer
		logger = log.New(&buf, "logger: ", log.Lshortfile)
	)

	p, err := ProvOGraph(v1, k, sha, urlloc)
	if err != nil {
		return err
	}

	//  Normalized sha only valid for JSON-LD  this is? only valid for JSON-LD
	// provsha, err := common.GetNormSHA(p, v1) // Moved to the normalized sha value
	provsha := common.GetSHA(p) // Moved to the normalized sha value
	// if err != nil {
	// 	logger.Printf("ERROR: URL: %s Action: Getting normalized sha  Error: %s\n", urlloc, err)
	// 	return err
	// }

	b := bytes.NewBufferString(p)

	objectName := fmt.Sprintf("prov/%s/%s.jsonld", k, provsha) // k is the name of the provider from config
	usermeta := make(map[string]string)                        // what do I want to know?
	usermeta["url"] = urlloc
	usermeta["sha1"] = sha // recall this is the sha of the about object, not the prov graph itself

	bucketName := "gleaner" //   fmt.Sprintf("gleaner-summoned/%s", k) // old was just k
	contentType := "application/ld+json"

	// Upload the file with FPutObject

	_, err = mc.PutObject(bucketName, objectName, b, int64(b.Len()), minio.PutObjectOptions{ContentType: contentType, UserMetadata: usermeta})
	if err != nil {
		logger.Printf("%s", objectName)
		logger.Fatalln(err) // Fatal?   seriously?    I guess this is the object write, so the run is likely a bust at this point, but this seems a bit much still.
	}

	return err
}

// ProvOGraph is a simpler provo prov function
// Against better judgment rather than build triples, I'll just
// template build them like with the nanoprov function
func ProvOGraph(v1 *viper.Viper, k, sha, urlloc string) (string, error) {

	// get the time
	currentTime := time.Now()
	date := fmt.Sprintf("%s", currentTime.Format("2006-01-02"))

	// open the config to get the runID later
	mcfg := v1.GetStringMapString("gleaner")

	// pull domains since we need to align k (stupid var for name here) with the PID value
	var domains []Sources
	err := v1.UnmarshalKey("sources", &domains)
	if err != nil {
		log.Println(err)
	}

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

	// build the struct to pass to the template parser

	// make the URN string // TODO make an extracted function to share with nabu
	bucketName := "gleaner"
	//objectURN := fmt.Sprintf("summoned:%s:%s", k, sha)
	objectURN := fmt.Sprintf("milled:%s:%s", k, sha)

	gp := fmt.Sprintf("urn:%s:%s", bucketName, objectURN)
	td := ProvData{RESID: urlloc, SHA256: sha, PID: pid, SOURCE: k, DATE: date, RUNID: mcfg["runid"], URN: gp, PNAME: pname, DOMAIN: domain}
	tmpl := provTplt()

	var doc bytes.Buffer
	t, err := template.New("prov").Parse(tmpl)
	if err != nil {
		panic(err)
	}
	err = t.Execute(&doc, td)
	if err != nil {
		panic(err)
	}

	// log.Print(doc.String())

	return doc.String(), err
}

func provTplt() string {

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

// NanoProvGraph generates a JSON-LD based nanopub prov graph for
// a resource collected.
func NanoProvGraph(k, sha, urlloc string) (string, error) {
	tmpl := nanoprov()

	currentTime := time.Now()
	date := fmt.Sprintf("%s", currentTime.Format("2006-01-02"))

	td := ProvData{RESID: urlloc, SHA256: sha, PID: "re3",
		SOURCE: k, DATE: date, RUNID: "testrunid"}

	var doc bytes.Buffer

	t, err := template.New("prov").Parse(tmpl)
	if err != nil {
		panic(err)
	}
	err = t.Execute(&doc, td)
	if err != nil {
		panic(err)
	}

	// fmt.Print(doc.String())

	return doc.String(), nil
}

func nanoprov() string {

	t := `{
  "@context": {
    "gleaner": "https://voc.gleaner.io/id/",
    "np": "http://www.nanopub.org/nschema#",
    "prov": "http://www.w3.org/ns/prov#",
    "xsd": "http://www.w3.org/2001/XMLSchema#"
  },
  "@set": [
    {
      "@id": "gleaner:nanopub/{{.SHA256}}",
      "@type": "np:NanoPublication",
      "np:hasAssertion": {
        "@id": "gleaner:nanopub/{{.SHA256}}#assertion"
      },
      "np:hasProvenance": {
        "@id": "gleaner:nanopub/{{.SHA256}}#provenance"
      },
      "np:hasPublicationInfo": {
        "@id": "gleaner:nanopub/{{.SHA256}}#pubInfo"
      }
    },
    {
      "@id": "gleaner:nanopub/{{.SHA256}}#assertion",
      "@graph": {
        "@id": "gleaner:{{.SHA256}}",
        "@type": "schema:Dataset",
        "identifier": [
          {
            "@type": "schema:PropertyValue",
            "name": "GraphSHA",
            "description": "A SHA256 sha stamp on the harvested data graph from a URL",
            "value": "{{.SHA256}}"
          },
          {
            "@type": "schema:PropertyValue",
            "name": "ProviderID",
            "description": "The id provided with the data graph by the provider",
            "value": "{{.PID}}"
          },
          {
            "@type": "schema:PropertyValue",
            "name": "URL",
            "description": "The URL harvested by gleaner",
            "value": "{{.RESID}}"
          }
        ]
      }
    },
    {
      "@id": "gleaner:nanopub/{{.SHA256}}#provenance",
      "@graph": {
        "@id": "gleaner:nanopub/{{.SHA256}}#assertion",
        "prov:wasGeneratedAtTime": {
          "@value": "{{.DATE}}",
          "@type": "xsd:dateTime"
        },
        "prov:wasDerivedFrom": {
          "@id": "URL of the resources and/or  @id from resource"
        },
        "prov:wasAttributedTo": {
          "@id": "Can I put the Institution base URl or ID here"
        }
      }
    },
    {
      "@id": "gleaner:nanopub/{{.SHA256}}#pubInfo",
      "@graph": {
        "@id": "gleaner:nanopub/{{.SHA256}}#nanopub",
        "prov:wasAttributedTo": {
          "@id": "gleaner:tool/gleaner"
        },
        "prov:generatedAtTime": {
          "@value": "2019-10-23T14:38:00Z",
          "@type": "xsd:dateTime"
        }
      }
    }
  ]
}
`
	return t

}
