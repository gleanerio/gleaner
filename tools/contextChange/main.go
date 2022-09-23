package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	//"github.com/gleanerio/gleaner/internal/summoner/acquire"
)

// Simple test of the nanoprov function
func main() {

	arg := os.Args[1]
	dat, err := os.ReadFile(arg)
	if err != nil {
		fmt.Println("Error reading file")
	}

	jld, err := fixContext(string(dat))
	//jld, err := fixContextString(string(dat))
	if err != nil {
		fmt.Println("Error reading file")
	}

	fmt.Println(jld)
	// pass to context change
}

// Our first json fixup in existence.
// If the top-level JSON-LD context is a string instead of an object,
// this function corrects it.
//func DEPRECATEDfixContextString(jsonld string) (string, error) {
//	var err error
//	jsonContext := gjson.Get(jsonld, "@context")
//
//	switch jsonContext.Value().(type) {
//	case string:
//		jsonld, err = sjson.Set(jsonld, "@context", map[string]interface{}{"@vocab": jsonContext.String()})
//	}
//	return jsonld, err
//}

// fixContext unifies and updates the context string altering.  It replaces both fixContextUrl and
func fixContext(jsonld string) (string, error) {
	var err error

	sdoc := "https://schema.org/" // eventually set in config so ignore always true/false conditional test below for now

	// grab the cotext
	c := gjson.Get(jsonld, "@context")

	// check to see if we can cast this to a map
	cm, ok := c.Value().(map[string]interface{})
	if !ok {

		// check to see if we can cast this to an []map
		acm, ok := c.Value().([]interface{})
		if !ok {
			// we are not a recognized context
			fmt.Println("This is not a recognized []map context either, will drop throgh to string check")
		} else {
			for x := range acm {
				e := acm[x].(map[string]interface{})
				fmt.Println(x)
				fmt.Println(len(acm))
				fmt.Println(e)
				//cm2, _ := acm[x].Value().(map[string]interface{}) // should an OK check
				for k, v := range e {
					fmt.Printf("Key: %s  Value: %s\n", k, v)

					// seed if v can do v.(string) and if not continue on..   don't deal with this gnis_url type things
					//"schema": "http://schema.org/",
					//	"NAME": "schema:name",
					//	"gnis_url": {
					//	"@id": "schema:subjectOf",
					//		"@type": "@id"
					//}

					if _, ok := v.(string); !ok {
						continue
					}

					if strings.HasPrefix(v.(string), "https://schema.org") {
						if v.(string) == sdoc {
							// we are good..  including trailing / as well, so leave
							return jsonld, nil
						} else {
							tns := fmt.Sprintf("@context.%s", k)
							jsonld, err = sjson.Set(jsonld, tns, sdoc)
							//return jsonld, err
						}
					} else if strings.HasPrefix(v.(string), "http://schema.org") {
						tns := ""
						if v.(string) == sdoc {
							// we are good..  including trailing / as well, so leave
							return jsonld, nil
						} else {
							if strings.HasPrefix(k, "@") {
								tns = fmt.Sprintf("@context.%d.\\%s", x, k) ////@context/@vocab
							} else {
								tns = fmt.Sprintf("@context.%d.%s", x, k) ////@context/@vocab
							}
							fmt.Printf("FIRST CHECK MARK:%s:%s\n", tns, sdoc)
							jsonld, err = sjson.Set(jsonld, tns, sdoc)

						}
					}
				}
			}
			return jsonld, err
		}

		fmt.Println("-----  string context -------")
		fmt.Println(c.Value().(string))
		// if not it's a string and we can just regex it..  let's not promote to map
		// check for https://schema.org/   (trailing / ?)  and fix
		if c.Value().(string) == "http://schema.org/" {
			if strings.Compare(sdoc, "http://schema.org/") == 0 {
				return jsonld, nil // all good, return
			} else {
				jsonld, err = sjson.Set(jsonld, "@context", sdoc)
				return jsonld, nil
			}
		} else if c.Value().(string) == "https://schema.org/" {
			if strings.Compare(sdoc, "https://schema.org/") == 0 {
				return jsonld, nil // all good, return
			} else {
				jsonld, err = sjson.Set(jsonld, "@context", sdoc)
				return jsonld, nil
			}
		}
		// repeat the above with the error of a missing trailing /
		// check for this second so we don't match the shorter substring pattern first, if the string
		// with the / is there we want to find it and leave this function first.
		if c.Value().(string) == "http://schema.org" {
			if strings.Compare(sdoc, "http://schema.org") == 0 {
				return jsonld, nil // all good, return
			} else {
				jsonld, err = sjson.Set(jsonld, "@context", sdoc)
				return jsonld, nil
			}
		} else if c.Value().(string) == "https://schema.org" {
			if strings.Compare(sdoc, "https://schema.org") == 0 {
				return jsonld, nil // all good, return
			} else {
				jsonld, err = sjson.Set(jsonld, "@context", sdoc)
				return jsonld, nil
			}
		}
	} else {
		for k, v := range cm {
			fmt.Printf("Key: %s  Value: %s\n", k, v)

			// if value is our schema.org issue then update this key
			// try to ready this for a "config"  option

			// NOTE we are checking the namespace as a prefix WITHOUT the training / to check
			// on it too in the logic
			if strings.HasPrefix(v.(string), "https://schema.org") {
				if v.(string) == sdoc {
					// we are good..  including trailing / as well, so leave
					return jsonld, nil
				} else {
					tns := fmt.Sprintf("@context.%s", k)
					jsonld, err = sjson.Set(jsonld, tns, sdoc)
					return jsonld, err
				}
			} else if strings.HasPrefix(v.(string), "http://schema.org") {
				tns := ""
				if v.(string) == sdoc {
					// we are good..  including trailing / as well, so leave
					return jsonld, nil
				} else {
					if strings.HasPrefix(k, "@") {
						tns = fmt.Sprintf("@context.\\%s", k) ////@context/@vocab
					} else {
						tns = fmt.Sprintf("@context.%s", k) ////@context/@vocab
					}
					fmt.Printf("CHECK MARK:%s:%s\n", tns, sdoc)
					jsonld, err = sjson.Set(jsonld, tns, sdoc)
					return jsonld, err
				}
			}
		}
	}

	// if we are still here, just return what we showed up with
	return jsonld, err
}
