package gon3

import (
	"crypto/sha1"
	"fmt"
	"sort"
)

// algorithm: http://www.hpl.hp.com/techreports/2001/HPL-2001-293.pdf
// also: http://blog.datagraph.org/2010/03/rdf-isomorphism

type CanonicalGraph struct {
	graph      *Graph
	nodes      []Term
	bNodes     []*BlankNode
	ungrounded map[*BlankNode]string
	grounded   map[Term]string
}

func (g *Graph) Canonicalize() *CanonicalGraph {
	cg := &CanonicalGraph{
		graph:      g,
		nodes:      g.NodesSorted(),
		ungrounded: make(map[*BlankNode]string),
		grounded:   make(map[Term]string),
	}
	for _, n := range cg.nodes {
		if isBlankNode(n) {
			cg.bNodes = append(cg.bNodes, n.(*BlankNode))
		}
	}
	return cg
}

func (cg1 *CanonicalGraph) IsomorphicTo(cg2 *CanonicalGraph) bool {
	// both graphs need same number of nodes
	if len(cg1.nodes) != len(cg2.nodes) {
		fmt.Println("Graphs not same size")
		return false
	}
	// make sure non-blank nodes match
	for i, n1 := range cg1.nodes {
		if !n1.Equals(cg2.nodes[i]) {
			fmt.Println("Graphs' non-blank nodes don't match")
			fmt.Printf("first: %+v\nsecond: %+v\n", cg1.nodes, cg2.nodes)
			return false
		}
	}
	// try to build a bijection between the graphs
	return cg1.BijectTo(cg2)
}

func (cg1 *CanonicalGraph) BijectTo(cg2 *CanonicalGraph) bool {
	// ground as many nodes as possible
	cg1.groundNodes()
	cg2.groundNodes()
	// ensure grounded nodes built at same rate
	if !cg1.verifyGrounded(cg2) {
		fmt.Println("verifyGrounded failed")
		return false
	}
	// map bnodes in cg1 to bnodes in cg2
	cg2UngroundedTmp := make(map[*BlankNode]string)
	for k, v := range cg2.ungrounded {
		cg2UngroundedTmp[k] = v
	}
	bijection := make(map[*BlankNode]*BlankNode)
	//fmt.Printf("ug1: %+v\nug2: %+v\n", cg1.ungrounded, cg2UngroundedTmp)
	for bn1, h1 := range cg1.ungrounded {
		for bn2, h2 := range cg2UngroundedTmp {
			if h2 == h1 {
				bijection[bn1] = bn2
				delete(cg2UngroundedTmp, bn2)
			}
		}
	}
	// if all nodes accounted for in mapping, success
	if cg1.validBijectionTo(bijection, cg2) {
		return true
	}

	// mark two ungrounded nodes with matching sigs as grounded and recurse
	for _, bn1 := range cg1.bNodes {
		if _, has := cg1.grounded[bn1]; has {
			fmt.Println("A")
			continue
		}
		for _, bn2 := range cg2.bNodes {
			if _, has := cg2.grounded[bn2]; has {
				fmt.Println("B")
				continue
			}
			if cg1.ungrounded[bn1] != cg2.ungrounded[bn2] {
				fmt.Println("C")
				continue
			}
			// try setting this pair as grounded
			hash := sha1.Sum([]byte(bn1.String()))
			cg1.grounded[bn1] = string(hash[:])
			cg2.grounded[bn2] = string(hash[:])
			fmt.Println("recursing")
			if cg1.BijectTo(cg2) {
				return true
			}
			// backtrack
			delete(cg1.grounded, bn1)
			delete(cg2.grounded, bn2)
		}
	}
	// if we have exhausted all signature matches, fail
	fmt.Println("signature matches exhausted")
	return false
}

func (cg1 *CanonicalGraph) validBijectionTo(bij map[*BlankNode]*BlankNode, cg2 *CanonicalGraph) bool {
	//fmt.Printf("bijmap: %+v\n", bij)
	nods1 := make([]Term, 0)
	nods2 := make([]Term, 0)
	cNods1 := make([]Term, 0)
	cNods2 := make([]Term, 0)
	for n1, n2 := range bij {
		nods1 = append(nods1, n1)
		nods2 = append(nods2, n2)
	}
	for _, n := range cg1.bNodes {
		cNods1 = append(cNods1, n)
	}
	for _, n := range cg2.bNodes {
		cNods2 = append(cNods2, n)
	}
	nodes1 := TermSlice(nods1)
	nodes2 := TermSlice(nods2)
	sort.Sort(nodes1)
	sort.Sort(nodes2)
	cNodes1 := TermSlice(cNods1)
	cNodes2 := TermSlice(cNods2)
	sort.Sort(cNodes1)
	sort.Sort(cNodes2)
	if len(nodes1) != len(cNodes1) || len(nodes2) != len(cNodes2) {
		fmt.Println("invalid bijection: lengths don't match")
		//fmt.Printf("%s\n%s\n%s\n%s\n", nodes1, cNodes1, nodes2, cNodes2)
		return false
	}
	for i := 0; i < len(nodes1); i++ {
		if nodes1[i] != cNodes1[i] {
			fmt.Println("invalid bijection: mismatched node")
			return false
		}
		if nodes2[i] != cNodes2[i] {
			fmt.Println("invalid bijection: mismatched node")
			return false
		}
	}
	return true
}

func (cg1 *CanonicalGraph) verifyGrounded(cg2 *CanonicalGraph) bool {
	for _, h1 := range cg1.grounded {
		found := false
		for _, h2 := range cg2.grounded {
			if h1 == h2 {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	for _, h2 := range cg2.grounded {
		found := false
		for _, h1 := range cg1.grounded {
			if h1 == h2 {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (cg *CanonicalGraph) groundNodes() {
	//fmt.Printf("START\ngrounded: %+v\nungrounded: %+v\n", cg.grounded, cg.ungrounded)
	for {
		// note current num of grounded nodes
		startLen := len(cg.grounded)
		// mark nodes as grounded by their membership in triples
		for _, n := range cg.bNodes {
			if _, has := cg.grounded[n]; !has {
				fmt.Println("hashing a node")
				cg.hashNode(n)
			}
		}
		// TODO: do we need to ground nodes with unique signatures?
		// break if no new nodes have been grounded
		if len(cg.grounded) == startLen {
			break
		}
	}
	//fmt.Printf("grounded: %+v\nungrounded: %+v\nEND\n", cg.grounded, cg.ungrounded)
}

func (cg *CanonicalGraph) hashNode(bn *BlankNode) {
	tripleSignatures := make([]string, 0)
	grounded := true
	for trip := range cg.graph.IterTriples() {
		if trip.includes(bn) {
			tripleSignatures = append(tripleSignatures, cg.getHashString(trip, bn))
			// if there are any other ungrounded blank nodes in this triple,
			// mark this blank node as ungrounded
			for term := range trip.IterNodes() {
				if isBlankNode(term) {
					bnod := term.(*BlankNode)
					_, present := cg.grounded[bnod]
					if !term.Equals(bn) && !present {
						grounded = false
					}
				}
			}
		}
	}
	sort.Strings(tripleSignatures)
	hash := sha1.Sum([]byte(fmt.Sprintf("%v", tripleSignatures)))
	if grounded {
		cg.grounded[bn] = string(hash[:])
	}
	cg.ungrounded[bn] = string(hash[:])
}

func (cg *CanonicalGraph) getHashString(t *Triple, n *BlankNode) string {
	str := ""
	for term := range t.IterNodes() {
		hash, grounded := cg.grounded[term]
		switch {
		case n == term:
			str += "itself"
		case grounded:
			str += hash
		case isBlankNode(n):
			str += "a blank node"
		default:
			str += n.String()
		}
	}
	return str
}
