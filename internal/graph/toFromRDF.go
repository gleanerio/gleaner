package graph

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"strings"

	"github.com/knakk/rdf"
	"github.com/piprate/json-gold/ld"
)

// fmt.Println("---------------nq -> json-ld------------------")
// jsonld, _ := nqToJSONLD(quads)

// fmt.Println("--------------json-ld -> nq----------------")
// nq, _ := jsonldToNQ(string(jsonld))

// fmt.Println("------------nq -> nt + context----------------")
// nt, g, _ := nqToNTCtx(nq)

// fmt.Println("------------nt + ctx -> nq----------------")
// nq2, _ := ntToNq(triples, "context")

func NtToNq(nt, ctx string) (string, error) {
	dec := rdf.NewTripleDecoder(strings.NewReader(nt), rdf.NTriples)
	tr, err := dec.DecodeAll()
	if err != nil {
		log.Printf("Error decoding triples: %v\n", err)
	}

	// loop on tr and make a set of quads
	var sa []string
	for i := range tr {
		q, err := makeQuad(tr[i], ctx)
		if err != nil {
			log.Println(err)
		}
		sa = append(sa, q)
		//fmt.Print(q)
	}

	return strings.Join(sa, ""), err
}

// makeQuad I pulled this from my ObjectEngine code in case I needed to
// use in the ntToNQ() function to add a context to each triple in turn.
// It may not be needed/used in this code
func makeQuad(t rdf.Triple, c string) (string, error) {
	newctx, err := rdf.NewIRI(c) // this should be  c
	ctx := rdf.Context(newctx)

	q := rdf.Quad{t, ctx}

	buf := bytes.NewBufferString("")

	qs := q.Serialize(rdf.NQuads)
	_, err = fmt.Fprintf(buf, "%s", qs)
	if err != nil {
		return "", err
	}

	return buf.String(), err
}

// NqToNTCtx  Converts quads to triples and return the graph name separately
func NqToNTCtx(inquads string) (string, string, error) {
	dec := rdf.NewQuadDecoder(strings.NewReader(inquads), rdf.NQuads)
	tr, err := dec.DecodeAll()
	if err != nil {
		log.Printf("Error decoding triples: %v\n", err)
	}

	// loop on tr and make a set of triples
	ntr := []rdf.Triple{}
	for i := range tr {
		ntr = append(ntr, tr[i].Triple)
	}

	// Assume context of first triple sis context of all triples
	// TODO..   this is stupid if not dangers, at least return []string of all the contexts
	// that were in the graph.
	ctx := tr[0].Ctx
	g := ctx.String()

	outtriples := ""
	buf := bytes.NewBufferString(outtriples)
	enc := rdf.NewTripleEncoder(buf, rdf.NTriples)
	err = enc.EncodeAll(ntr)
	if err != nil {
		log.Printf("Error encoding triples: %v\n", err)
	}
	err = enc.Close()
	if err != nil {
		return "", "", err
	}

	tb := bytes.NewBuffer([]byte(""))
	for k := range ntr {
		tb.WriteString(ntr[k].Serialize(rdf.NTriples))
	}

	return tb.String(), g, err
}

// NqToJSONLD convert quads to JSON-LD?  (or is this NT...  see next func)
// Am I losing the context here (ie, no graph[]:)
func NqToJSONLD(triples string) ([]byte, error) {
	proc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")
	// add the processing mode explicitly if you need JSON-LD 1.1 features
	options.ProcessingMode = ld.JsonLd_1_1

	doc, err := proc.FromRDF(triples, options)
	if err != nil {
		panic(err)
	}

	// ld.PrintDocument("JSON-LD output", doc)
	b, err := json.MarshalIndent(doc, "", " ")

	return b, err
}

func JsonldToNQ(jsonld string) (string, error) {
	proc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")
	// add the processing mode explicitly if you need JSON-LD 1.1 features
	options.ProcessingMode = ld.JsonLd_1_1
	options.Format = "application/n-quads"

	var myInterface interface{}
	err := json.Unmarshal([]byte(jsonld), &myInterface)
	if err != nil {
		log.Println("Error when transforming JSON-LD document to interface:", err)
		return "", err
	}

	triples, err := proc.ToRDF(myInterface, options) // returns triples but toss them, just validating
	if err != nil {
		log.Println("Error when transforming JSON-LD document to RDF:", err)
		return "", err
	}

	return fmt.Sprintf("%v", triples), err
}
