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
	gd := buildGraph(entries)
	// i, err := writeRDF(gd, bucketname)
	i, err := millerutils.WriteRDF(gd, fmt.Sprintf("%s_prov", bucketname))

	if err != nil {
		log.Println(err)
	}

	log.Printf("Write prov record for %s with len %d\n", bucketname, i)
}

func buildGraph(pi []utils.Entry) string {

	// make UUID here to make the baseuri uniqie
	// uuid. .Init()
	u := uuid.NewV4()

	// Set a base URI
	baseURI := fmt.Sprintf("https://provisium.io/id/%s", u)
	g := rdf2go.NewGraph(baseURI)

	// r is of type io.Reader
	bt, ot := baseTriples("Label  here", "OrgName", "OrgMailto", u.String(), baseURI)
	err := g.Parse(strings.NewReader(bt), "text/turtle")
	err = g.Parse(strings.NewReader(ot), "text/turtle")
	if err != nil {
		log.Println(err)
	}

	// Add in the members of the prov:Collection
	for item := range pi {
		triple1 := rdf2go.NewTriple(rdf2go.NewResource(fmt.Sprintf("http://provisium.io/id//%spagecollection", u.String())), rdf2go.NewResource("prov:hadMember"), rdf2go.NewResource(pi[item].Urlval))
		g.Add(triple1)
	}

	// Dump graph contents to NTriples
	out := g.String()

	return out

}

func baseTriples(Label, Name, MailTo, pid, baseURI string) (string, string) {

	// Would be nice to have a URL here for them too..  maybe other data as well
	orgtriples := fmt.Sprintf(`@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
@prefix xsd: <http://www.w3.org/2001/XMLSchema#> .
@prefix foaf: <http://xmlns.com/foaf/0.1/> .
@prefix prov: <http://www.w3.org/ns/prov#> .
@prefix eos: <http://esipfed.org/prov/eos#> .
@prefix dcat: <https://www.w3.org/ns/dcat> .
@prefix rdfs: <http://www.w3.org/2000/01/rdf-schema#> .
@prefix : <http://provisium.io/id/%s/> .

:datafacility
    a prov:Agent, prov:Organization ;
    rdfs:label "%s"^^xsd:string ;
    foaf:givenName "%s" ;
    foaf:mbox <mailto:%s> .
	`, pid, Label, Name, MailTo)

	bt := fmt.Sprintf(`@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
@prefix xsd: <http://www.w3.org/2001/XMLSchema#> .
@prefix foaf: <http://xmlns.com/foaf/0.1/> .
@prefix prov: <http://www.w3.org/ns/prov#> .
@prefix eos: <http://esipfed.org/prov/eos#> .
@prefix dcat: <https://www.w3.org/ns/dcat> .
@prefix rdfs: <http://www.w3.org/2000/01/rdf-schema#> .
@prefix : <http://provisium.io/id/%s/> .

# Will need to honor and deference this URI to a landing page for this prov data
<%s>
    a prov:Bundle, prov:Entity ;
    rdfs:label "A collection of provenance related to the creation of a P418 index"^^xsd:string ;
    prov:generatedAtTime "2018-02-14T00:00:00Z"^^xsd:dateTime ;
    prov:wasAttributedTo :processingActivity1 .

:esso
    a prov:Agent, prov:Organization ;
    rdfs:label "EarthCube Science Support Office"^^xsd:string ;
    foaf:givenName "USGS" ;
    # need URL
    foaf:mbox <mailto:info@earthcube.org> .

:processingCode
    a eos:software, prov:Entity ;
    rdfs:label "EarthCube Project 418 Indexer"^^xsd:string ;
    # what voc to use to link to software repo?  (other ID?)  just need a URl for now
    prov:wasAttributedTo :esso .

:dataset
    a eos:product, prov:Entity ;
    rdfs:label "Dataset included spatial, text and graph results from the activity"^^xsd:string ;
	prov:wasAttributedTo :esso ;
	prov:wasDerivedFrom :pagecollection ;  
    prov:wasGeneratedBy :processingActivity1 .

:processingActivity1
    a eos:processStep, prov:Activity ;
    rdfs:label "Generation of indexes (spatial, text, graph) from the processed pages"^^xsd:string ;
    prov:endedAtTime "2011-07-14T02:02:02Z"^^xsd:dateTime ;
    prov:startedAtTime "2011-07-14T01:01:01Z"^^xsd:dateTime ;
    prov:used :processingCode ;  
    prov:used :processingCode, :pagecollection ;
	prov:wasAssociatedWith :esso .
	
:pagecollection 
	rdfs:label "URIs submitted to the pingback service"^^xsd:string; 
	prov:wasAttributedTo :datafacility ;
	a prov:Collection .   
	
`, pid, baseURI)

	return bt, orgtriples
}
