package common

import (
	"fmt"
	configTypes "github.com/gleanerio/gleaner/internal/config"
	"os"
	"path/filepath"
)

/*
This is to test various identifier
It uses a structure of expectations to run a series of individual tests with the name: testname_jsonfilename.

In the future, the JSON should probably be loaded from a file in resources_test folder.
*/

/* info on possible packages:
https://cburgmer.github.io/json-path-comparison/
using https://github.com/ohler55/ojg

test your jsonpaths here:
http://jsonpath.herokuapp.com/
There are four implementations... so you can see if one might be a little quirky
*/
import (
	"bytes"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

// jsonexpectations is in test_common_structs

// testdata is in internal/common/testdata/identifier
// thoughts are that these many be migrated to  an Approval Test approach.
// gets rid of the extpectations part, and would match the entire returned identifier object.

// should record a table of the file sha and normalize triple sha for each file

var empty = []configTypes.Sources{}

// using idenfiters as a stand in for array of identifiers.

func testValidJsonPath(tests []jsonexpectations, t *testing.T) {
	for _, test := range tests {
		for i, json := range test.json {
			t.Run(fmt.Sprint(test.name, "_", i), func(t *testing.T) {
				if test.ignore {
					return
				}
				path := filepath.Join("testdata", "identifier", json)
				assert.FileExistsf(t, path, "Datafile Missing: {path}")
				source, err := os.ReadFile(path)
				if err != nil {
					t.Fatal("error reading source file:", err)
				}

				result, err := GetIdentifierByPath(test.IdentifierPaths, string(source))
				valStr := fmt.Sprint(result)
				assert.Equal(t, test.expected, valStr)
				assert.Nil(t, err)
			})
		}
	}

	//t.Run("@id", func(t *testing.T) {
	//
	//	result, err := GetIdentifierByPath(sources[0].IdentifierPath, jsonId)
	//	valStr := fmt.Sprint(result)
	//	assert.Equal(t, "[idenfitier]", valStr)
	//	assert.Nil(t, err)
	//})
	//t.Run(".idenfitier", func(t *testing.T) {
	//
	//	result, err := GetIdentifierByPath(sources[1].IdentifierPath, jsonId)
	//	valStr := fmt.Sprint(result)
	//	assert.Equal(t, "[doi:10.1575/1912/bco-dmo.2343.1]", valStr)
	//	assert.Nil(t, err)
	//})
	//t.Run("$.idenfitier", func(t *testing.T) {
	//
	//	result, err := GetIdentifierByPath(sources[2].IdentifierPath, jsonId)
	//	valStr := fmt.Sprint(result)
	//	assert.Equal(t, "[doi:10.1575/1912/bco-dmo.2343.1]", valStr)
	//	assert.Nil(t, err)
	//})
	// to do: test for valid JSON but invalid RDF triples
}

// test the array paths
func testValidJsonPaths(tests []jsonexpectations, t *testing.T) {
	for _, test := range tests {
		for i, json := range test.json {
			t.Run(fmt.Sprint(test.name, "_", i), func(t *testing.T) {
				if test.ignore {
					return
				}
				path := filepath.Join("testdata", "identifier", json)
				assert.FileExistsf(t, path, "Datafile Missing: {path}")
				source, err := os.ReadFile(path)
				if err != nil {
					t.Fatal("error reading source file:", err)
				}
				paths := strings.Split(test.IdentifierPaths, ",")
				result, foundPath, err := GetIdentiferByPaths(paths, string(source))
				valStr := fmt.Sprint(result)
				assert.Equal(t, test.expected, valStr, "expected Failed")
				assert.Equal(t, test.expectedPath, foundPath, "matched Path Failed")
				assert.Nil(t, err)
			})
		}

	}

}

/*
this tests a single path against a single json file
*/
func TestValidJsonPathInput(t *testing.T) {

	var tests = []jsonexpectations{
		// default
		{
			name:          "@id",
			json:          map[string]string{"jsonID": "jsonId.json"},
			errorExpected: false,

			IdentifierPaths: `$['@id']`,
			expected:        "[idenfitier]",
			expectedPath:    "$['@id']",
			ignore:          false,
		},
		//https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/earthchem2.json
		{
			name:            "@.identifier",
			json:            map[string]string{"jsonID": "jsonId.json"},
			errorExpected:   false,
			IdentifierPaths: "@.identifier",
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1]",
			expectedPath:    "@.identifier",
			ignore:          false,
		},
		//https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/earthchem2.json
		{
			name:            "$.identifier",
			json:            map[string]string{"jsonID": "jsonId.json"},
			errorExpected:   false,
			IdentifierPaths: "$.identifier",
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1]",
			expectedPath:    "$.identifier",
			ignore:          false,
		},
		// argo example: https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/argo.json
		{
			name:            "identifiers Array ",
			json:            map[string]string{"jsonID": "jsonId.json"},
			errorExpected:   false,
			IdentifierPaths: "$.identifierSArray[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value",
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1N doi:10.1575/1912/bco-dmo.2343.1]",
			expectedPath:    "$.identifierSArray[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value",
			ignore:          false,
		},
		{
			name:          "identifier_obj",
			json:          map[string]string{"jsonID": "jsonId.json"},
			errorExpected: false,
			//	IdentifierPath: "$.identifierObj[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value",
			//IdentifierPath: "$.identifierObj.propertyID[@=='https://registry.identifiers.org/registry/doi')]",
			IdentifierPaths: "$.identifierObj.value",
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1]",
			expectedPath:    "$.identifierObj.value",
			ignore:          false,
		},
		//https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/earthchem2.json
		// this will not work since the || does not work
		{
			name:            " identifier or id",
			json:            map[string]string{"jsonID": "jsonId.json"},
			errorExpected:   false,
			IdentifierPaths: "[ $.identifiers[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value || $.['@id'] ]",
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1]",
			expectedPath:    "[ $.identifiers[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value || $.['@id'] ]",
			ignore:          true,
		},
		// identifier as array: https://github.com/earthcube/GeoCODES-Metadata/blob/main/metadata/Dataset/allgood/bcodmo1.json
		/* needs work
		"identifier": [

		       {
		           "@type": "PropertyValue",
		           "@id": "https://doi.org/10.1575/1912/bco-dmo.2343.1",
		           "propertyID": "https://registry.identifiers.org/registry/doi",
		           "value": "doi:10.1575/1912/bco-dmo.2343.1",
		           "url": "https://doi.org/10.1575/1912/bco-dmo.2343.1"
		       }
		   ],
		*/
		// this does not work fancy array index issues. Would be nice
		{
			name:          "identifierSArray slice",
			json:          map[string]string{"jsonID": "jsonId.json"},
			errorExpected: false,
			//IdentifierPath: "$.identifierSArray[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value[-1:]",
			//IdentifierPaths: []string{"$.identifierSArray[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value.[-1:]"},
			IdentifierPaths: "$.identifierSArray[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value[0]",
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1]",
			expectedPath:    "$.identifierSArray[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value.[0]",
			ignore:          true,
		},
	}

	testValidJsonPath(tests, t)
}

func TestValidJsonPathsInput(t *testing.T) {

	var tests = []jsonexpectations{
		// default
		// should work for all
		{
			name: "@id",
			json: map[string]string{"jsonID": "jsonIdPaths.json", "jsonIdentifier": "jsonIdentifierPath.json",
				"jsonobjectId":                "jsonIdentifierObjectPath.json",
				"jsonIdentifierArraySingle":   "jsonIdentifierArraySingle.json",
				"jsonIdentifierArrayMultiple": "jsonIdentifierArrayMultiple.json",
			},
			errorExpected: false,

			IdentifierPaths: `$['@id']`,
			expected:        "[idenfitier]",
			expectedPath:    "$['@id']",
			ignore:          false,
		},
		//https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/earthchem2.json
		// this returns an empty set [] https://cburgmer.github.io/json-path-comparison/results/dot_notation_on_object_without_key.html
		{
			name: "$.identifier.$id",
			//json:            []string{jsonId},
			json: map[string]string{"jsonID": "jsonIdPaths.json"}, //"jsonIdentifier": jsonIdentifier,
			//"jsonobjectId": jsonIdentifierObject,
			//"jsonIdentifierArraySingle": jsonIdentifierArraySingle,
			//"jsonIdentifierArrayMultiple": jsonIdentifierArrayMultiple,

			errorExpected:   false,
			IdentifierPaths: "$.identifier.value,$.identifier,$['@id']",
			expected:        "[idenfitier]",
			expectedPath:    "$['@id']",
			ignore:          false,
		},
		{
			name: "$.identifier.$.identifier",
			//json:            []string{jsonIdentifier},
			json:            map[string]string{"jsonIdentifier": "jsonIdentifierPath.json"},
			errorExpected:   false,
			IdentifierPaths: "$.identifier.value,$.identifier,$['@id']",
			expected:        "[doi:10]",
			expectedPath:    "$.identifier",
			ignore:          false,
		},
		//https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/earthchem2.json
		{
			name: "$.identifierObjBracket",
			//json:            []string{jsonIdentifierObject},
			json: map[string]string{
				"jsonobjectId": "jsonIdentifierObjectPath.json",
			},
			errorExpected:   false,
			IdentifierPaths: "$.identifier['value'],$.identifier,$['@id']",
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1]",
			expectedPath:    "$.identifier['value']",
			ignore:          false,
		},
		{
			name: "$.identifierObjDot",
			//json:            []string{jsonIdentifierObject},
			json: map[string]string{
				"jsonobjectId": "jsonIdentifierObjectPath.json",
			},
			errorExpected:   false,
			IdentifierPaths: "$.identifier.value,$.identifier,$['@id']",
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1]",
			expectedPath:    "$.identifier.value",
			ignore:          false,
		},
		{
			name: "$.identifierObjCheck",
			//json:            []string{jsonIdentifierObject},
			json: map[string]string{
				"jsonobjectId": "jsonIdentifierObjectPath.json",
			},
			errorExpected:   false,
			IdentifierPaths: "$.identifier.value,$.identifier,$['@id']",
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1]",
			expectedPath:    "$.identifier.value",
			ignore:          false,
		},
		//https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/earthchem2.json
		{
			name: "@.identifierArraySimple",
			//json:            []string{jsonIdentifierArraySingle},
			json: map[string]string{
				"jsonIdentifierArraySingle": "jsonIdentifierArraySingle.json",
			},
			errorExpected:   false,
			IdentifierPaths: "$.identifier[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value,$.identifier.value,$.identifier.$['@id']",
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1]",
			expectedPath:    "$.identifier[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value",
			ignore:          false,
		},

		//https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/earthchem2.json
		{
			name: "@.identifierArrayMultiple",
			//json:            []string{jsonIdentifierArrayMultiple},
			json: map[string]string{
				"jsonIdentifierArrayMultiple": "jsonIdentifierArrayMultiple.json",
			},
			errorExpected:   false,
			IdentifierPaths: "$.identifier[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value,$.identifier.value,$.identifier,$['@id']",
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1N doi:10.1575/1912/bco-dmo.2343.1]",
			expectedPath:    "$.identifier[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value",
			ignore:          false,
		},
		{
			name: "@.identifierProblemChildIris",
			//json:            []string{jsonIdentifierArrayMultiple},
			json: map[string]string{
				"problem child": "problemChildIris.json",
			},
			errorExpected:   false,
			IdentifierPaths: "$.identifier[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value,$.identifier.value,$.identifier,$['@id']",
			expected:        "[https://ds.iris.edu/ds/products/emtf/]",
			expectedPath:    "$['@id']",
			ignore:          false,
		},
		{
			name: "@.identifierProblemChildOpenTopo",
			//json:            []string{jsonIdentifierArrayMultiple},
			json: map[string]string{
				"problem child opentopo": "problemChildOpentop.json",
			},
			errorExpected:   false,
			IdentifierPaths: "$.identifier[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value,$.identifier.value,$.identifier,$['@id']",
			expected:        "[OTDS.062020.32611.1]",
			expectedPath:    "$.identifier.value",
			ignore:          false,
		},
	}
	testValidJsonPaths(tests, t)
}

func testGenerateJsonPathIdentifier(tests []jsonexpectations, t *testing.T) {

	//mock configre file
	// paths are relative to the code
	var vipercontext = []byte(`
context:
  cache: true
contextmaps:
- file: ../../configs/schemaorg-current-https.jsonld
  prefix: https://schema.org/
- file: ../../configs/schemaorg-current-https.jsonld
  prefix: http://schema.org/
sources:
- sourcetype: sitemap
  name: test
  logo: https://opentopography.org/sites/opentopography.org/files/ot_transp_logo_2.png
  url: https://opentopography.org/sitemap.xml
  headless: false
  pid: https://www.re3data.org/repository/r3d100010655
  propername: OpenTopography
  domain: http://www.opentopography.org/
  active: false
  credentialsfile: ""
  other: {}
  headlesswait: 0
  delay: 0
  IdentifierType: identifiersha
`)

	for _, test := range tests {
		for i, json := range test.json {
			// needs to be defiend in the loop, so that each run has it's own configuration.
			// otherwise changing the sources information in a multi-threaded ent has issues
			viperVal := viper.New()
			viperVal.SetConfigType("yaml")
			viperVal.ReadConfig(bytes.NewBuffer(vipercontext))
			sources, err := configTypes.GetSources(viperVal)

			if err != nil {
				assert.Fail(t, err.Error())
			}

			s := sources[0]
			s.IdentifierType = test.IdentifierType
			s.IdentifierPath = test.IdentifierPaths
			t.Run(fmt.Sprint(test.name, "_", i), func(t *testing.T) {
				if test.ignore {
					return
				}
				path := filepath.Join("testdata", "identifier", json)
				assert.FileExistsf(t, path, "Datafile Missing: {path}")
				source, err := os.ReadFile(path)
				if err != nil {
					t.Fatal("error reading source file:", err)
				}
				result, err := GenerateIdentifier(viperVal, s, string(source))
				//valStr := fmt.Sprint(result.uniqueId)
				assert.Equal(t, test.expected, result.UniqueId, "uuid faild")
				assert.Equal(t, test.expectedPath, result.MatchedPath, "matched path failed")
				assert.Equal(t, test.IdentifierType, result.IdentifierType, "identifier failed")
				assert.Nil(t, err)
			})
		}
	}
}
func testGenerateFileShaIdentifier(tests []jsonexpectations, t *testing.T) {

	//mock configre file
	// paths are relative to the code
	var vipercontext = []byte(`
context:
  cache: true
contextmaps:
- file: ../../configs/schemaorg-current-https.jsonld
  prefix: https://schema.org/
- file: ../../configs/schemaorg-current-https.jsonld
  prefix: http://schema.org/
sources:
- sourcetype: sitemap
  name: test
  logo: https://opentopography.org/sites/opentopography.org/files/ot_transp_logo_2.png
  url: https://opentopography.org/sitemap.xml
  headless: false
  pid: https://www.re3data.org/repository/r3d100010655
  propername: OpenTopography
  domain: http://www.opentopography.org/
  active: false
  credentialsfile: ""
  other: {}
  headlesswait: 0
  delay: 0
  IdentifierType: filesha
`)

	for _, test := range tests {
		for i, json := range test.json {
			// needs to be defiend in the loop, so that each run has it's own configuration.
			// otherwise changing the sources information in a multi-threaded ent has issues
			viperVal := viper.New()
			viperVal.SetConfigType("yaml")
			viperVal.ReadConfig(bytes.NewBuffer(vipercontext))
			sources, err := configTypes.GetSources(viperVal)

			if err != nil {
				assert.Fail(t, err.Error())
			}

			s := sources[0]
			s.IdentifierType = test.IdentifierType
			s.IdentifierPath = test.IdentifierPaths
			t.Run(fmt.Sprint(test.name, "_", i), func(t *testing.T) {
				if test.ignore {
					return
				}
				path := filepath.Join("testdata", "identifier", json)
				assert.FileExistsf(t, path, "Datafile Missing: {path}")
				source, err := os.ReadFile(path)
				if err != nil {
					t.Fatal("error reading source file:", err)
				}
				result, err := GenerateIdentifier(viperVal, s, string(source))
				//valStr := fmt.Sprint(result.uniqueId)
				assert.Equal(t, test.expected, result.UniqueId, "uuid failed")
				assert.Equal(t, test.expectedPath, result.MatchedPath, "matched path failed")
				assert.Equal(t, test.IdentifierType, result.IdentifierType, "identifiertype match failed")
				assert.Nil(t, err)
			})
		}
	}
}

func TestGenerateFileShaIdentifier(t *testing.T) {

	var tests = []jsonexpectations{
		// default
		// should work for all
		{
			name: "jsonsha_array",
			json: map[string]string{
				"jsonIdentifierArrayMultiple": "jsonIdentifierArrayMultiple.json",
			},
			errorExpected:   false,
			IdentifierType:  configTypes.JsonSha,
			IdentifierPaths: "$['@id']",
			expected:        "7bff4b860c6df6f12f408955d0e78da2dea9e268",
			expectedPath:    "",
			ignore:          false,
		},
		{
			name: "normalizedsha_array",
			json: map[string]string{
				"jsonIdentifierArrayMultiple": "jsonIdentifierArrayMultiple.json",
			},
			errorExpected:   false,
			IdentifierType:  configTypes.NormalizedJsonSha,
			IdentifierPaths: "$['@id']",
			expected:        "37626666346238363063366466366631326634303839353564306537386461326465613965323638",
			expectedPath:    "",
			ignore:          false,
		},
		{
			name: "normalizedsha_id",
			json: map[string]string{
				"jsonidentifier": "jsonIdentifierPath.json",
			},
			errorExpected:   false,
			IdentifierType:  configTypes.NormalizedJsonSha,
			IdentifierPaths: "$['@id']",
			expected:        "38646664383435363837333837653337663236383132343335313436613363343462376231346262",
			expectedPath:    "",
			ignore:          false,
		},
		{
			name: "jsonsha_id",
			json: map[string]string{
				"jsonidentifier": "jsonIdentifierPath.json",
			},
			errorExpected:   false,
			IdentifierType:  configTypes.JsonSha,
			IdentifierPaths: "$['@id']",
			expected:        "8dfd845687387e37f26812435146a3c44b7b14bb",
			expectedPath:    "",
			ignore:          false,
		},
	}

	testGenerateFileShaIdentifier(tests, t)
}

func TestGenerateJsonPathIdentifier(t *testing.T) {

	var tests = []jsonexpectations{
		// default
		// should work for all

		{
			name: "@id_first",
			json: map[string]string{
				"jsonIdentifierArrayMultiple": "jsonIdentifierArrayMultiple.json",
			},
			errorExpected:   false,
			IdentifierType:  configTypes.IdentifierSha,
			IdentifierPaths: "$['@id']",
			expected:        "0fe143f05d6dbff260874a9a6e8da77243c74db0",
			expectedPath:    "$['@id']",
			ignore:          false,
		},
		{
			name: "identifier_default_path",
			json: map[string]string{
				"jsonIdentifierArrayMultiple": "jsonIdentifierArrayMultiple.json",
			},
			errorExpected:   false,
			IdentifierType:  configTypes.IdentifierSha,
			IdentifierPaths: "",
			expected:        "e59f7f11a5615bcee6f35c92d8a2162e5b611944",
			expectedPath:    "$.identifier[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value",
			ignore:          false,
		},
	}

	testGenerateJsonPathIdentifier(tests, t)
}

func TestValidJsonPathGraphInput(t *testing.T) {

	var tests = []jsonexpectations{
		// default

		{
			name:          "identifieGraph Not Graph",
			json:          map[string]string{"jsonID": "jsonIdentifierArrayMultiple.json"},
			errorExpected: true,
			//IdentifierPath: "$.identifierSArray[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value[-1:]",
			IdentifierPaths: "$['@graph'][?(@['@type']=='schema:Dataset')]['@id'],$.identifier[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value,$.identifier.value,$.identifier,$['@id']",

			expected:     "[doi:10.1575/1912/bco-dmo.2343.1N doi:10.1575/1912/bco-dmo.2343.1]",
			expectedPath: "$.identifier[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value",
			ignore:       false,
		},
		// grr. Ugly since the herokuapp no longer runs: used this a hint, then raw debugging: https://cburgmer.github.io/json-path-comparison/

		// this one $['@graph]*[?(@['@type']=='schema:Dataset')]  gives false here: https://jsonpath.curiousconcept.com/
		// $['@graph']*.['@type'] returns types
		// $['@graph'].*.@id returns types
		//$.@graph*[?(@.@type=="schema:Dataset")] false bad when debuggin. cannot start with an @

		// workslocally:
		// returns nil: "$['@graph']*[?(@['@type']=='schema:Dataset')]"
		// returns full object: "$['@graph'][?(@['@type']=='schema:Dataset')]"
		// returns @id: "$['@graph'][?(@['@type']=='schema:Dataset')]['@id']"  fails at: https://jsonpath.curiousconcept.com/
		{
			name:          "identifiersGraph",
			json:          map[string]string{"jsonGraph": "jsonGraphWifire.json"},
			errorExpected: false,
			//IdentifierPath: "$.identifierSArray[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value[-1:]",
			IdentifierPaths: "$['@graph'][?(@['@type']=='schema:Dataset')]['@id']",

			expected:     "[https://wifire-data.sdsc.edu/dataset/8fd44c38-f6d3-429c-a785-1498dfaa2a6a]",
			expectedPath: "$['@graph'][?(@['@type']=='schema:Dataset')]['@id']",
			ignore:       false,
		},
		{
			name:          "identifiersGraphLong",
			json:          map[string]string{"jsonGraph": "jsonGraphWifire.json"},
			errorExpected: false,
			//IdentifierPath: "$.identifierSArray[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value[-1:]",
			IdentifierPaths: "$['@graph'][?(@['@type']=='schema:Dataset')]['@id'],$.identifier[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value,$.identifier.value,$.identifier,$['@id']",

			expected:     "[https://wifire-data.sdsc.edu/dataset/8fd44c38-f6d3-429c-a785-1498dfaa2a6a]",
			expectedPath: "$['@graph'][?(@['@type']=='schema:Dataset')]['@id']",
			ignore:       false,
		},
	}

	testValidJsonPaths(tests, t)
}
