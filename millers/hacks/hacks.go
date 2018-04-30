package hacks

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

// IEDA1 is set to remove <Crossref Funder ID:  from strings in the IEDA graph
func IEDA1(input string) string {
	return strings.Replace(input, "<Crossref Funder ID:", "<", -1)
}

// Neotoma1 removes [] lists in schema.org/contentUrl which do not translate correctly into RDF
// They are valid JSON-LD, being part of the data mode.  They are not in the RDF data model,
// but rather in the vocabulary
// This should be resolvable by setting @container to @list in the context for this item...
// Need    "contentUrl":[ "http://api.neotomadb.org/v1/data/downloads/4294", "http://api.neotomadb.org/v1/data/downloads/9104",
// "http://api.neotomadb.org/v1/data/downloads/4295", "http://api.neotomadb.org/v1/data/downloads/9105", "http://api.neotomadb.org/v1/data/downloads/4296"],
func Neotoma1(input string) string {
	scanner := bufio.NewScanner(strings.NewReader(input))
	var buffer bytes.Buffer
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "<http://schema.org/contentUrl>") {
			if strings.Contains(scanner.Text(), ",") {
				pa := strings.SplitAfter(scanner.Text(), "<http://schema.org/contentUrl> ")
				pa1 := pa[1]
				pa1 = strings.TrimLeft(pa1, "<")
				pa1 = strings.TrimRight(pa1, "> .")
				urls := strings.Split(pa1, ",")
				ot := strings.Split(scanner.Text(), " ") //get original triples,  will fault on later space....
				for i := range urls {
					buffer.WriteString(fmt.Sprintf("%s %s <%s> .\n", ot[0], ot[1], strings.TrimSpace(urls[i])))
				}
			} else {
				buffer.WriteString(scanner.Text())
			}
		}
	}
	return buffer.String()
}
