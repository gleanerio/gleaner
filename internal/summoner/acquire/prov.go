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

// StoreProv creates and stores a prov record for each JSON-LD data graph
// k is the domain / provider
// sha is the sha of the JSON-LD file summoned
// urlloc is the URL for the resource (source URL)
func StoreProv(v1 *viper.Viper, mc *minio.Client, k, sha, urlloc string) error {
	// read config file
	//miniocfg := v1.GetStringMapString("minio")
	//bucketName := miniocfg["bucket"] //   get the top level bucket for all of gleaner operations from config file
	bucketName, err := configTypes.GetBucketName(v1)
	//var (
	//	buf    bytes.Buffer
	//	logger = log.New(&buf, "logger: ", log.Lshortfile)
	//)

	p, err := provOGraph(v1, k, sha, urlloc, "milled") // TODO default to milled till I update the rest of the code and remove this version of the function

	// NOTE
	// Setting the value of the prov reference is a connection to how Nabu and the queries related to prov work.  When Nabu loads from summoned, the
	// graph value is set and this needs to match what is here.  Else, we load from milled and the same connection has to take place.   Loading from
	// milled is hard when we are dealing with large sitegraphs.  ??

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

	contentType := "application/ld+json"

	// Upload the file with FPutObject
	_, err = mc.PutObject(context.Background(), bucketName, objectName, b, int64(b.Len()), minio.PutObjectOptions{ContentType: contentType, UserMetadata: usermeta})
	if err != nil {
		log.Fatal(objectName, err)
		// Fatal?   seriously?    I guess this is the object write, so the run is likely a bust at this point, but this seems a bit much still.
	}

	return err
}

func StoreProvNG(v1 *viper.Viper, mc *minio.Client, k, sha, urlloc, objprefix string) error {
	// read config file
	//miniocfg := v1.GetStringMapString("minio")
	//bucketName := miniocfg["bucket"] //   get the top level bucket for all of gleaner operations from config file
	bucketName, err := configTypes.GetBucketName(v1)
	//var (
	//	buf    bytes.Buffer
	//	logger = log.New(&buf, "logger: ", log.Lshortfile)
	//)

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

	contentType := "application/ld+json"

	// Upload the file with FPutObject
	_, err = mc.PutObject(context.Background(), bucketName, objectName, b, int64(b.Len()), minio.PutObjectOptions{ContentType: contentType, UserMetadata: usermeta})
	if err != nil {
		log.Fatal(objectName, err)
		// Fatal?   seriously?    I guess this is the object write, so the run is likely a bust at this point, but this seems a bit much still.
	}

	return err
}

// provOGraph is a simpler provo prov function
// Against better judgment rather than build triples, I'll just
// template build them like with the nanoprov function
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

	// build the struct to pass to the template parser
	gp := fmt.Sprintf("urn:%s:%s", bucketName, sha)
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

// // NanoProvGraph generates a JSON-LD based nanopub prov graph for
// // a resource collected.
// func NanoProvGraph(k, sha, urlloc string) (string, error) {
// 	tmpl := nanoprov()

// 	currentTime := time.Now()
// 	date := fmt.Sprintf("%s", currentTime.Format("2006-01-02"))

// 	td := ProvData{RESID: urlloc, SHA256: sha, PID: "re3",
// 		SOURCE: k, DATE: date, RUNID: "testrunid"}

// 	var doc bytes.Buffer

// 	t, err := template.New("prov").Parse(tmpl)
// 	if err != nil {
// 		panic(err)
// 	}
// 	err = t.Execute(&doc, td)
// 	if err != nil {
// 		panic(err)
// 	}

// 	// fmt.Print(doc.String())

// 	return doc.String(), nil
// }

// func nanoprov() string {

// 	t := `{
//   "@context": {
//     "gleaner": "https://voc.gleaner.io/id/",
//     "np": "http://www.nanopub.org/nschema#",
//     "prov": "http://www.w3.org/ns/prov#",
//     "xsd": "http://www.w3.org/2001/XMLSchema#"
//   },
//   "@set": [
//     {
//       "@id": "gleaner:nanopub/{{.SHA256}}",
//       "@type": "np:NanoPublication",
//       "np:hasAssertion": {
//         "@id": "gleaner:nanopub/{{.SHA256}}#assertion"
//       },
//       "np:hasProvenance": {
//         "@id": "gleaner:nanopub/{{.SHA256}}#provenance"
//       },
//       "np:hasPublicationInfo": {
//         "@id": "gleaner:nanopub/{{.SHA256}}#pubInfo"
//       }
//     },
//     {
//       "@id": "gleaner:nanopub/{{.SHA256}}#assertion",
//       "@graph": {
//         "@id": "gleaner:{{.SHA256}}",
//         "@type": "schema:Dataset",
//         "identifier": [
//           {
//             "@type": "schema:PropertyValue",
//             "name": "GraphSHA",
//             "description": "A SHA256 sha stamp on the harvested data graph from a URL",
//             "value": "{{.SHA256}}"
//           },
//           {
//             "@type": "schema:PropertyValue",
//             "name": "ProviderID",
//             "description": "The id provided with the data graph by the provider",
//             "value": "{{.PID}}"
//           },
//           {
//             "@type": "schema:PropertyValue",
//             "name": "URL",
//             "description": "The URL harvested by gleaner",
//             "value": "{{.RESID}}"
//           }
//         ]
//       }
//     },
//     {
//       "@id": "gleaner:nanopub/{{.SHA256}}#provenance",
//       "@graph": {
//         "@id": "gleaner:nanopub/{{.SHA256}}#assertion",
//         "prov:wasGeneratedAtTime": {
//           "@value": "{{.DATE}}",
//           "@type": "xsd:dateTime"
//         },
//         "prov:wasDerivedFrom": {
//           "@id": "URL of the resources and/or  @id from resource"
//         },
//         "prov:wasAttributedTo": {
//           "@id": "Can I put the Institution base URl or ID here"
//         }
//       }
//     },
//     {
//       "@id": "gleaner:nanopub/{{.SHA256}}#pubInfo",
//       "@graph": {
//         "@id": "gleaner:nanopub/{{.SHA256}}#nanopub",
//         "prov:wasAttributedTo": {
//           "@id": "gleaner:tool/gleaner"
//         },
//         "prov:generatedAtTime": {
//           "@value": "2019-10-23T14:38:00Z",
//           "@type": "xsd:dateTime"
//         }
//       }
//     }
//   ]
// }
// `
// 	return t

// }

// StoreProv creates and stores a prov record for each JSON-LD data graph
// k is the domain / provider
// sha is the sha of the JSON-LD file summoned
// urlloc is the URL for the resource (source URL)
// func StoreProv(v1 *viper.Viper, mc *minio.Client, k, sha, urlloc string) error {
// 	// read config file
// 	miniocfg := v1.GetStringMapString("minio")
// 	bucketName := miniocfg["bucket"] //   get the top level bucket for all of gleaner operations from config file

// 	var (
// 		buf    bytes.Buffer
// 		logger = log.New(&buf, "logger: ", log.Lshortfile)
// 	)

// 	p, err := provOGraph(v1, k, sha, urlloc, "milled") // TODO default to milled till I update the rest of the code and remove this version of the function
// 	if err != nil {
// 		return err
// 	}

// 	// NOTE
// 	// Setting the value of the prov reference is a connection to how Nabu and the queries related to prov work.  When Nabu loads from summoned, the
// 	// graph value is set and this needs to match what is here.  Else, we load from milled and the same connection has to take place.   Loading from
// 	// milled is hard when we are dealing with large sitegraphs.  ??

// 	// Moved to the normalized sha value since normalized sha only valid for JSON-LD
// 	provsha := common.GetSHA(p)
// 	// provsha, err := common.GetNormSHA(p, v1) // Moved to the normalized sha value
// 	// if err != nil {
// 	// 	log.Printf("ERROR: URL: %s Action: Getting normalized sha  Error: %s\n", urlloc, err)
// 	// 	return err
// 	// }

// 	b := bytes.NewBufferString(p)

// 	objectName := fmt.Sprintf("prov/%s/%s.jsonld", k, provsha) // k is the name of the provider from config
// 	usermeta := make(map[string]string)                        // what do I want to know?
// 	usermeta["url"] = urlloc
// 	usermeta["sha1"] = sha // recall this is the sha of the about object, not the prov graph itself

// 	contentType := "application/ld+json"

// 	// Upload the file with FPutObject
// 	_, err = mc.PutObject(context.Background(), bucketName, objectName, b, int64(b.Len()), minio.PutObjectOptions{ContentType: contentType, UserMetadata: usermeta})
// 	if err != nil {
// 		log.Printf("%s", objectName)
// 		logger.Fatalln(err) // Fatal?   seriously?    I guess this is the object write, so the run is likely a bust at this point, but this seems a bit much still.
// 	}

// 	return err
// }
