package hacks

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

// IEDA1 is set to remove <Crossref Funder ID:
func IEDA1(input string) string {
	return strings.Replace(input, "<Crossref Funder ID:", "<", -1)
}

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
					buffer.WriteString(fmt.Sprintf("%s %s <%s>\n", ot[0], ot[1], strings.TrimSpace(urls[i])))
				}
			} else {
				buffer.WriteString(scanner.Text())
			}
		}
	}
	return buffer.String()
}
