package millerprov

import (
	"fmt"
	"log"
	"strings"

	"earthcube.org/Project418/gleaner/internal/common"
	"earthcube.org/Project418/gleaner/internal/millers/millerutils"
	"earthcube.org/Project418/gleaner/pkg/utils"

	"github.com/deiu/rdf2go"
	minio "github.com/minio/minio-go"
	"github.com/tidwall/gjson"
	"github.com/twinj/uuid"
)

// MockObjects test a concurrent version of calling mock
func MockObjects(mc *minio.Client, bucketname string, cs utils.Config) {
	entries := common.GetMillObjects(mc, bucketname)
	gd := buildGraph(entries, bucketname, cs)

	// write to S3
	i, err := millerutils.LoadToMinio(gd, "gleaner-milled", fmt.Sprintf("%s_prov.n3", bucketname), mc)
	if err != nil {
		log.Println(err)
	}

	// // write to file
	// i, err := millerutils.WriteRDF(gd, fmt.Sprintf("%s_prov", bucketname))
	// if err != nil {
	// 	log.Println(err)
	// }

	log.Printf("Wrote prov record for %s with len %d\n", bucketname, i)
}

func buildGraph(pi []common.Entry, bucketname string, cs utils.Config) string {
	// make UUID here to make the baseuri unique
	u := uuid.NewV4()

	// Set a base URI
	// baseURI := fmt.Sprintf("https://provisium.org/id/%s#", u.String())
	g := rdf2go.NewGraph("")

	// r is of type io.Reader
	ot := orgTriples(bucketname, u.String(), cs) // TODO..  bucketname obviously not what we want to do here....
	bt := baseTriples(u.String(), cs)            // TODO..  bucketname obviously not what we want to do here....

	err := g.Parse(strings.NewReader(bt), "text/turtle") // should be ntriples
	if err != nil {
		log.Println(err)
	}

	err = g.Parse(strings.NewReader(ot), "text/turtle")
	if err != nil {
		log.Println(err)
	}

	// TODO   why use a UUID here is the page collection is for a facility?  Just use label?
	// or use a UUID for different prov graphs?  then I need to keep them unique  (which is likely the case)
	// Add in the members of the prov:Collection
	for item := range pi {
		triple1 := rdf2go.NewTriple(rdf2go.NewResource(fmt.Sprintf("http://provisium.org/id/%s/pagecollection", u.String())), rdf2go.NewResource("http://www.w3.org/ns/prov#hadMember"), rdf2go.NewResource(pi[item].Urlval))
		g.Add(triple1)
		// reach inside the JSON-LD and pull out some other IDs.  Would prefer to do this with framing
		// but performance issues with so much context round tripping.  (no context cache yet)
		// altid1 :=
		v1 := gjson.Get(pi[item].Jld, "@id") // get id
		triple2 := rdf2go.NewTriple(rdf2go.NewResource(fmt.Sprintf("http://provisium.org/id/%s/pagecollection", u.String())), rdf2go.NewResource("http://www.w3.org/ns/prov#hadMember"), rdf2go.NewResource(v1.String()))
		g.Add(triple2)
	}

	// Dump graph contents to NTriples
	out := g.String()

	return out
}

// This should have the name of the organization exposing the data
func orgTriples(bname, u string, cs utils.Config) string {
	g := rdf2go.NewGraph("")
	s := fmt.Sprintf("http://provisium.org/id/%s/pagecollection", u)

	// pull url and shortname from cs based on Lable (aka  name)
	var shortname, url string
	for _, v := range cs.Sources {
		if bname == v.Name {
			shortname = v.ShortName
			url = v.URL
		}
	}

	// 	 grep related ocdProvP418.n3
	// 	<https://provisium.org/id/43989f37-8504-46af-a340-16641c9b5b0f#http://provisium.org/id/opencoredataorg/pagecollection> <https://provisium.org/id/43989f37-8504-46af-a340-16641c9b5b0f#http://www.w3.org/2000/01/rdf-schema#label> "A collection of provenance related to the creation of a P418 index" .
	// 	<https://provisium.org/id/43989f37-8504-46af-a340-16641c9b5b0f#http://provisium.org/id/opencoredataorg/pagecollection> <https://provisium.org/id/43989f37-8504-46af-a340-16641c9b5b0f#http://www.w3.org/2004/02/skos/core#related> <https://provisium.org/id/43989f37-8504-46af-a340-16641c9b5b0f#http://opencoredata.org/sitemap.xml> .

	// <https://provisium.org/id/43989f37-8504-46af-a340-16641c9b5b0f#http://provisium.org/id/opencoredataorg/pagecollection> <https://provisium.org/id/43989f37-8504-46af-a340-16641c9b5b0f#http://www.w3.org/2004/02/skos/core#related> <https://provisium.org/id/43989f37-8504-46af-a340-16641c9b5b0f#opencore> .

	if shortname != "" {
		g.Add(rdf2go.NewTriple(rdf2go.NewResource(s), rdf2go.NewResource("http://www.w3.org/2004/02/skos/core#related"), rdf2go.NewLiteral(shortname)))
	}

	if url != "" {
		g.Add(rdf2go.NewTriple(rdf2go.NewResource(s), rdf2go.NewResource("http://www.w3.org/2000/01/rdf-schema#seeAlso"), rdf2go.NewLiteral(url)))
	}

	// Build the triples
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s), rdf2go.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type"), rdf2go.NewResource("http://www.w3.org/ns/prov#hadMember")))
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s), rdf2go.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type"), rdf2go.NewResource("http://www.w3.org/ns/prov#Organization")))
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s), rdf2go.NewResource("http://www.w3.org/2000/01/rdf-schema#label"), rdf2go.NewLiteral(bname)))
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s), rdf2go.NewResource("http://xmlns.com/foaf/0.1/givenName"), rdf2go.NewLiteral(bname)))

	return g.String()
}

func baseTriples(u string, cs utils.Config) string {
	g := rdf2go.NewGraph("")

	// Bundel and entity
	s := fmt.Sprintf("http://provisium.org/id/%s/pagecollection", u)
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s), rdf2go.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type"), rdf2go.NewResource("http://www.w3.org/ns/prov#Bundle")))
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s), rdf2go.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type"), rdf2go.NewResource("http://www.w3.org/ns/prov#Entity")))
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s), rdf2go.NewResource("http://www.w3.org/2000/01/rdf-schema#label"), rdf2go.NewLiteral("A collection of provenance related to the creation of a P418 index")))
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s), rdf2go.NewResource("http://www.w3.org/ns/prov#generatedAtTime"), rdf2go.NewLiteral("input time here")))
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s), rdf2go.NewResource("http://www.w3.org/ns/prov#wasAttributedTo"), rdf2go.NewResource("http://provisium.org/processingActivity/ID"))) // TODO..  fix ID ??

	// Agent and org
	s1 := "http://provisium.org/datafacility/esso"
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s1), rdf2go.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type"), rdf2go.NewResource("http://www.w3.org/ns/prov#Agent")))
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s1), rdf2go.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type"), rdf2go.NewResource("http://www.w3.org/ns/prov#Organization")))
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s1), rdf2go.NewResource("http://www.w3.org/2000/01/rdf-schema#label"), rdf2go.NewLiteral("EarthCube Science Support Office")))
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s1), rdf2go.NewResource("http://xmlns.com/foaf/0.1/givenName"), rdf2go.NewLiteral("ESSO")))
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s1), rdf2go.NewResource("http://xmlns.com/foaf/0.1/mbox"), rdf2go.NewLiteral("info@earthcube.org")))

	// eos:software and entity
	s2 := "http://provisium.org/datafacility/processingCode/gleaner"
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s2), rdf2go.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type"), rdf2go.NewResource("http://esipfed.org/prov/eos#software")))
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s2), rdf2go.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type"), rdf2go.NewResource("http://www.w3.org/ns/prov#Entity")))
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s2), rdf2go.NewResource("http://www.w3.org/2000/01/rdf-schema#label"), rdf2go.NewLiteral("EarthCube Project 418 Indexer")))
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s2), rdf2go.NewResource("http://www.w3.org/ns/prov#wasAttributedTo"), rdf2go.NewResource("http://provisium.org/datafacility/esso")))

	s3 := "http://provisium.org/dataset/ID"
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s3), rdf2go.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type"), rdf2go.NewResource("http://esipfed.org/prov/eos#product")))
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s3), rdf2go.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type"), rdf2go.NewResource("http://www.w3.org/ns/prov#Entity")))
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s3), rdf2go.NewResource("http://www.w3.org/2000/01/rdf-schema#label"), rdf2go.NewLiteral("Dataset included spatial, text and graph results from the activity")))
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s3), rdf2go.NewResource("http://www.w3.org/ns/prov#wasAttributedTo"), rdf2go.NewResource("http://provisium.org/datafacility/esso")))
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s3), rdf2go.NewResource("http://www.w3.org/ns/prov#wasDerivedFrom"), rdf2go.NewResource("http://provisium.org/pagecollection/ID")))
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s3), rdf2go.NewResource("http://www.w3.org/ns/prov#wasGeneratedBy"), rdf2go.NewResource("http://provisium.org/processingActivity/ID")))

	s4 := "http://provisium.org/processingActivity/ID"
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s4), rdf2go.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type"), rdf2go.NewResource("http://esipfed.org/prov/eos#processStep")))
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s4), rdf2go.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type"), rdf2go.NewResource("http://www.w3.org/ns/prov#Activity")))
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s4), rdf2go.NewResource("http://www.w3.org/2000/01/rdf-schema#label"), rdf2go.NewLiteral("Generation of indexes (spatial, text, graph) from the processed pages")))
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s3), rdf2go.NewResource("http://www.w3.org/ns/prov#endedAtTime"), rdf2go.NewResource("http://provisium.org/processingActivity/ID")))
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s3), rdf2go.NewResource("http://www.w3.org/ns/prov#startedAtTime"), rdf2go.NewResource("http://provisium.org/processingActivity/ID")))
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s3), rdf2go.NewResource("http://www.w3.org/ns/prov#used"), rdf2go.NewResource("http://provisium.org/datafacility/processingCode/gleaner")))
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s3), rdf2go.NewResource("http://www.w3.org/ns/prov#used"), rdf2go.NewResource("http://provisium.org/pagecollection/ID")))
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s3), rdf2go.NewResource("http://www.w3.org/ns/prov#wasAssociatedWith"), rdf2go.NewResource("http://provisium.org/datafacility/esso")))

	s5 := "http://provisium.org/pagecollection/ID"
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s5), rdf2go.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type"), rdf2go.NewResource("http://www.w3.org/ns/prov#Collection")))
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s5), rdf2go.NewResource("http://www.w3.org/2000/01/rdf-schema#label"), rdf2go.NewLiteral("URIs accessed by gleaner")))
	g.Add(rdf2go.NewTriple(rdf2go.NewResource(s5), rdf2go.NewResource("http://www.w3.org/ns/prov#wasAttributedTo"), rdf2go.NewResource("http://provisium.org/datafacility/esso")))

	return g.String()

	// --------  old version below  ------------

	// 	bt := fmt.Sprintf(`@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
	// @prefix xsd: <http://www.w3.org/2001/XMLSchema#> .
	// @prefix foaf: <http://xmlns.com/foaf/0.1/> .
	// @prefix prov: <http://www.w3.org/ns/prov#> .
	// @prefix eos: <http://esipfed.org/prov/eos#> .
	// @prefix dcat: <https://www.w3.org/ns/dcat> .
	// @prefix rdfs: <http://www.w3.org/2000/01/rdf-schema#> .

	// # Will need to honor and deference this URI to a landing page for this prov data
	// <%s>
	//     a prov:Bundle, prov:Entity ;
	//     rdfs:label "A collection of provenance related to the creation of a P418 index"^^xsd:string ;
	//     prov:generatedAtTime "2018-02-14T00:00:00Z"^^xsd:dateTime ;
	//     prov:wasAttributedTo <http://provisium.org/processingActivity/ID>	.

	// <http://provisium.org/datafacility/esso>
	//     a prov:Agent, prov:Organization ;
	//     rdfs:label "EarthCube Science Support Office"^^xsd:string ;
	//     foaf:givenName "ESSO" ;
	//     # need URL
	//     foaf:mbox <mailto:info@earthcube.org> .

	// <http://provisium.org/datafacility/processingCode/gleaner>
	//     a eos:software, prov:Entity ;
	//     rdfs:label "EarthCube Project 418 Indexer"^^xsd:string ;
	//     # what voc to use to link to software repo?  (other ID?)  just need a URl for now
	//     prov:wasAttributedTo <http://provisium.org/datafacility/esso> .

	// <http://provisium.org/dataset/ID>
	//     a eos:product, prov:Entity ;
	//     rdfs:label "Dataset included spatial, text and graph results from the activity"^^xsd:string ;
	// 	prov:wasAttributedTo <http://provisium.org/datafacility/esso> ;
	// 	prov:wasDerivedFrom <http://provisium.org/pagecollection/ID>  ;
	//     prov:wasGeneratedBy <http://provisium.org/processingActivity/ID>	.

	// <http://provisium.org/processingActivity/ID>
	//     a eos:processStep, prov:Activity ;
	//     rdfs:label "Generation of indexes (spatial, text, graph) from the processed pages"^^xsd:string ;
	//     prov:endedAtTime "2011-07-14T02:02:02Z"^^xsd:dateTime ;
	//     prov:startedAtTime "2011-07-14T01:01:01Z"^^xsd:dateTime ;
	//     prov:used <http://provisium.org/datafacility/processingCode/gleaner>;
	//     prov:used <http://provisium.org/datafacility/processingCode/gleaner>,<http://provisium.org/pagecollection/ID>  ;
	// 	prov:wasAssociatedWith  <http://provisium.org/datafacility/esso> .

	// <http://provisium.org/pagecollection/ID>
	// 	rdfs:label "URIs accessed by gleaner"^^xsd:string;
	// 	prov:wasAttributedTo <http://provisium.org/datafacility/esso>	;
	// 	a prov:Collection .
	// `, baseURI)

	// 	return bt
}
