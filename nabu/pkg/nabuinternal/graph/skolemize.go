package graph

import (
	"bufio"
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"strings"

	"github.com/rs/xid"
)

// Skolemization replaces blank nodes with URIs  The mapping approach is needed since this
// function can be used on a whole data graph, not just a single triple
// reference: https://www.w3.org/TR/rdf11-concepts/#dfn-skolem-iri
func Skolemization(nq, key string) (string, error) {
	scanner := bufio.NewScanner(strings.NewReader(nq))

	// need for long lines like in Internet of Water
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	// since a data graph may have several references to any given blank node, we need to keep a
	// map of our update.  It is also why the ID needs a non content approach since the blank node will
	// be in a different triple set from time to time and we can not ensure what order we might encounter them at.
	m := make(map[string]string) // make a map here to hold our updated strings

	for scanner.Scan() {
		split := strings.Split(scanner.Text(), " ")
		sold := split[0]
		oold := split[2]

		if strings.HasPrefix(sold, "_:") { // we are a blank node
			if _, ok := m[sold]; ok { // fmt.Printf("We had %s, already\n", sold)
			} else {
				guid := xid.New()
				snew := fmt.Sprintf("<https://gleaner.io/xid/genid/%s>", guid.String())
				m[sold] = snew
			}
		}

		// scan the object nodes too.. though we should find nothing here.. the above wouldn't find
		if strings.HasPrefix(oold, "_:") { // we are a blank node
			// check map to see if we have this in our value already
			if _, ok := m[oold]; ok {
				// fmt.Printf("We had %s, already\n", oold)
			} else {
				guid := xid.New()
				onew := fmt.Sprintf("<https://gleaner.io/xid/genid/%s>", guid.String())
				m[oold] = onew
			}
		}
	}

	var err = scanner.Err()
	if err != nil {
		log.Errorf("Error decoding source: %s\n", key)
		log.Error(err)
	}

	filebytes := []byte(nq)

	for k, v := range m {
		//fmt.Printf("Replace %s with %v \n", k, v)
		// The +" " is need since we have to avoid
		// _:b1 replacing _:b13 with ...3
		filebytes = bytes.Replace(filebytes, []byte(k+" "), []byte(v+" "), -1)
	}

	return string(filebytes), err
}
