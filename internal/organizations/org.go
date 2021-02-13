package organizations

import (
	"bytes"
	"fmt"
	"log"
	"text/template"

	"github.com/earthcubearchitecture-project418/gleaner/internal/summoner/acquire"

	"github.com/minio/minio-go"
	"github.com/spf13/viper"
)

// BuildGraph makes a graph from the Gleaner config file source
// load this to a /sources bucket (change this to sources naming convention?)
func BuildGraph(mc *minio.Client, v1 *viper.Viper) {
	var (
		buf    bytes.Buffer
		logger = log.New(&buf, "logger: ", log.Lshortfile)
	)

	log.Print("Building organization graph from config file")

	var domains []acquire.Sources
	err := v1.UnmarshalKey("sources", &domains)
	if err != nil {
		log.Println(err)
	}

	// Sources: Name, Logo, URL, Headless, Pid

	for k := range domains {

		jld, err := orggraph(domains[k])
		if err != nil {
			log.Println(err)
		}
		// log.Print(jld)

		b := bytes.NewBufferString(jld)

		// orgsha := common.GetSHA(jld)
		// objectName := fmt.Sprintf("orgs/%s/%s.nq", domains[k].Name, orgsha) // k is the name of the provider from config
		objectName := fmt.Sprintf("orgs/%s.nq", domains[k].Name) // k is the name of the provider from config
		bucketName := "gleaner"                                  //   fmt.Sprintf("gleaner-summoned/%s", k) // old was just k
		contentType := "application/ld+json"

		// Upload the file with FPutObject

		_, err = mc.PutObject(bucketName, objectName, b, int64(b.Len()), minio.PutObjectOptions{ContentType: contentType})
		if err != nil {
			logger.Printf("%s", objectName)
			logger.Fatalln(err) // Fatal?   seriously?    I guess this is the object write, so the run is likely a bust at this point, but this seems a bit much still.
		}

		// send to org graph function
		// write to minio prov bucket

	}

}

func orggraph(k acquire.Sources) (string, error) {

	tmpl := orgTemplate()

	var doc bytes.Buffer
	t, err := template.New("prov").Parse(tmpl)
	if err != nil {
		log.Println(err)
	}
	err = t.Execute(&doc, k)
	if err != nil {
		log.Println(err)
	}

	return doc.String(), err

}

func orgTemplate() string {

	t := `{
		"@context": {
			"@vocab": "https://schema.org/"
		},
		"@id": "https://gleaner.io/id/org/{{.Name}}",
		"@type": "Organization",
		"url": "{{.URL}}",
		"name": "{{.Name}}",
		 "identifier": {
			"@type": "PropertyValue",
			"propertyID": "https://registry.identifiers.org/registry/doi",
			"url": "{{.PID}}",
			"description": "Persistent identifier for this organization"
		}
	}
	`

	return t
}
