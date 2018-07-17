package millerprov

import (
	"earthcube.org/Project418/gleaner/millers/millerutils"
	//	"bytes"

	"fmt"
	"log"
	"strings"
	//	"log"

	"earthcube.org/Project418/gleaner/utils"
	"github.com/deiu/rdf2go"
	minio "github.com/minio/minio-go"
	"github.com/twinj/uuid"
)

// MockObjects test a concurrent version of calling mock
func MockObjects(mc *minio.Client, bucketname string) {
	entries := utils.GetMillObjects(mc, bucketname)
	gd := buildGraph(entries, bucketname)
	// i, err := writeRDF(gd, bucketname)
	i, err := millerutils.WriteRDF(gd, fmt.Sprintf("%s_prov", bucketname))

	if err != nil {
		log.Println(err)
	}

	log.Printf("Write prov record for %s with len %d\n", bucketname, i)
}

func buildGraph(pi []utils.Entry, bucketname string) string {
	// make UUID here to make the baseuri uniqie
	// uuid. .Init()
	u := uuid.NewV4() // just make a unique ID for the base URI for this graph??

	// cf := "./config.json"
	// c := utils.LoadConfiguration(&cf)
	// now build a search util to locate needed info at this point

	// Set a base URI
	baseURI := fmt.Sprintf("https://provisium.io/id/%s#", u.String())
	g := rdf2go.NewGraph(baseURI)

	// r is of type io.Reader
	bt, ot := baseTriples(bucketname, bucketname, u.String(), baseURI) // TODO..  bucketname obviously not what we want to do here....

	err := g.Parse(strings.NewReader(bt), "text/turtle")
	if err != nil {
		log.Println(err)
	}

	err = g.Parse(strings.NewReader(ot), "text/turtle")
	if err != nil {
		log.Println(err)
	}

	// Add in the members of the prov:Collection
	for item := range pi {
		triple1 := rdf2go.NewTriple(rdf2go.NewResource(fmt.Sprintf("http://provisium.io/id/%s/pagecollection", u.String())), rdf2go.NewResource("prov:hadMember"), rdf2go.NewResource(pi[item].Urlval))
		g.Add(triple1)
	}

	// Dump graph contents to NTriples
	out := g.String()

	return out
}

func baseTriples(Label, Name, pid, baseURI string) (string, string) {

	// Would be nice to have a URL here for them too..  maybe other data as well
	orgtriples := fmt.Sprintf(`@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
@prefix xsd: <http://www.w3.org/2001/XMLSchema#> .
@prefix foaf: <http://xmlns.com/foaf/0.1/> .
@prefix prov: <http://www.w3.org/ns/prov#> .
@prefix eos: <http://esipfed.org/prov/eos#> .
@prefix dcat: <https://www.w3.org/ns/dcat> .
@prefix rdfs: <http://www.w3.org/2000/01/rdf-schema#> .

<http://provisium.org/datafacility/%s>
    a prov:Agent, prov:Organization ;
    rdfs:label "%s"^^xsd:string ;
    foaf:givenName "%s" .
	`, Label, Label, Name)

	bt := fmt.Sprintf(`@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
@prefix xsd: <http://www.w3.org/2001/XMLSchema#> .
@prefix foaf: <http://xmlns.com/foaf/0.1/> .
@prefix prov: <http://www.w3.org/ns/prov#> .
@prefix eos: <http://esipfed.org/prov/eos#> .
@prefix dcat: <https://www.w3.org/ns/dcat> .
@prefix rdfs: <http://www.w3.org/2000/01/rdf-schema#> .



# Will need to honor and deference this URI to a landing page for this prov data
<%s>
    a prov:Bundle, prov:Entity ;
    rdfs:label "A collection of provenance related to the creation of a P418 index"^^xsd:string ;
    prov:generatedAtTime "2018-02-14T00:00:00Z"^^xsd:dateTime ;
    prov:wasAttributedTo <http://provisium.org/processingActivity/ID>	.

<http://provisium.org/datafacility/esso>
    a prov:Agent, prov:Organization ;
    rdfs:label "EarthCube Science Support Office"^^xsd:string ;
    foaf:givenName "ESSO" ;
    # need URL
    foaf:mbox <mailto:info@earthcube.org> .

<http://provisium.org/datafacility/processingCode/gleaner>
    a eos:software, prov:Entity ;
    rdfs:label "EarthCube Project 418 Indexer"^^xsd:string ;
    # what voc to use to link to software repo?  (other ID?)  just need a URl for now
    prov:wasAttributedTo <http://provisium.org/datafacility/esso> .

<http://provisium.org/dataset/ID>
    a eos:product, prov:Entity ;
    rdfs:label "Dataset included spatial, text and graph results from the activity"^^xsd:string ;
	prov:wasAttributedTo <http://provisium.org/datafacility/esso> ;
	prov:wasDerivedFrom <http://provisium.org/pagecollection/ID>  ;  
    prov:wasGeneratedBy <http://provisium.org/processingActivity/ID>	.

<http://provisium.org/processingActivity/ID>
    a eos:processStep, prov:Activity ;
    rdfs:label "Generation of indexes (spatial, text, graph) from the processed pages"^^xsd:string ;
    prov:endedAtTime "2011-07-14T02:02:02Z"^^xsd:dateTime ;
    prov:startedAtTime "2011-07-14T01:01:01Z"^^xsd:dateTime ;
    prov:used <http://provisium.org/datafacility/processingCode/gleaner>;  
    prov:used <http://provisium.org/datafacility/processingCode/gleaner>,<http://provisium.org/pagecollection/ID>  ;
	prov:wasAssociatedWith  <http://provisium.org/datafacility/esso> .
	
<http://provisium.org/pagecollection/ID> 
	rdfs:label "URIs accessed by gleaner"^^xsd:string; 
	prov:wasAttributedTo <http://provisium.org/datafacility/esso>	;
	a prov:Collection .   
	
`, baseURI)

	return bt, orgtriples
}
