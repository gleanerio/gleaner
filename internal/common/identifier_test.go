package common

import (
	"bytes"
	"fmt"
	configTypes "github.com/gleanerio/gleaner/internal/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

type expectations struct {
	name            string
	json            map[string]string
	IdentifierType  string `default:filesha`
	IdentifierPaths []string
	expected        string
	expectedPath    string
	errorExpected   bool `default:false`
	ignore          bool `default:false`
}

var empty = []configTypes.Sources{}

// using idenfiters as a stand in for array of identifiers.

func testValidJsonPath(tests []expectations, t *testing.T) {
	for _, test := range tests {
		for i, json := range test.json {
			t.Run(fmt.Sprint(test.name, "_", i), func(t *testing.T) {
				if test.ignore {
					return
				}
				result, err := GetIdentifierByPath(test.IdentifierPaths[0], json)
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
func testValidJsonPaths(tests []expectations, t *testing.T) {
	for _, test := range tests {
		for i, json := range test.json {
			t.Run(fmt.Sprint(test.name, "_", i), func(t *testing.T) {
				if test.ignore {
					return
				}
				result, foundPath, err := GetIdentiferByPaths(test.IdentifierPaths, json)
				valStr := fmt.Sprint(result)
				assert.Equal(t, test.expected, valStr)
				assert.Equal(t, test.expectedPath, foundPath)
				assert.Nil(t, err)
			})
		}

	}

}

/*
this tests a single path against a single json file
*/
func TestValidJsonPathInput(t *testing.T) {

	var jsonId = `{
"@id":"idenfitier",
"identifier":"doi:10.1575/1912/bco-dmo.2343.1",
"identifierArray": [	
	{
	"@type": "PropertyValue",
	"@id": "https://doi.org/10.1575/1912/bco-dmo.2343.1",
	"propertyID": "https://registry.identifiers.org/registry/doi",
	"value": "doi:10.1575/1912/bco-dmo.2343.1",
	"url": "https://doi.org/10.1575/1912/bco-dmo.2343.1"
	}
],
"identifierSArray": [	
	{
	"@type": "PropertyValue",
	"@id": "https://doi.org/10.1575/1912/bco-dmo.2343.1",
	"propertyID": "https://registry.identifiers.org/registry/doi",
	"value": "doi:10.1575/1912/bco-dmo.2343.1",
	"url": "https://doi.org/10.1575/1912/bco-dmo.2343.1"
	},
	{
	"@type": "PropertyValue",
	"@id": "https://doi.org/10.1575/1912/bco-dmo.2343.N",
	"propertyID": "https://registry.identifiers.org/registry/doi",
	"value": "doi:10.1575/1912/bco-dmo.2343.1N",
	"url": "https://doi.org/10.1575/1912/bco-dmo.2343.N"
	},
	{
	"@type": "PropertyValue",
	"@id": "https://doi.org/10.1575/1912/bco-dmo.2343.P",
	"propertyID": "https://purl.org",
	"value": "doi:10.1575/1912/bco-dmo.2343.P",
	"url": "https://doi.org/10.1575/1912/bco-dmo.2343.P"
	}
],
"identifierObj": 
	{
	"@type": "PropertyValue",
	"@id": "https://doi.org/10.1575/1912/bco-dmo.2343.1",
	"propertyID": "https://registry.identifiers.org/registry/doi",
	"value": "doi:10.1575/1912/bco-dmo.2343.1",
	"url": "https://doi.org/10.1575/1912/bco-dmo.2343.1"
	}

}`
	var tests = []expectations{
		// default
		{
			name:          "@id",
			json:          map[string]string{"jsonID": jsonId},
			errorExpected: false,

			IdentifierPaths: []string{`$['@id']`},
			expected:        "[idenfitier]",
			expectedPath:    "$['@id']",
			ignore:          false,
		},
		//https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/earthchem2.json
		{
			name:            "@.identifier",
			json:            map[string]string{"jsonID": jsonId},
			errorExpected:   false,
			IdentifierPaths: []string{"@.identifier"},
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1]",
			expectedPath:    "@.identifier",
			ignore:          false,
		},
		//https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/earthchem2.json
		{
			name:            "$.identifier",
			json:            map[string]string{"jsonID": jsonId},
			errorExpected:   false,
			IdentifierPaths: []string{"$.identifier"},
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1]",
			expectedPath:    "$.identifier",
			ignore:          false,
		},
		// argo example: https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/argo.json
		{
			name:            "identifiers Array ",
			json:            map[string]string{"jsonID": jsonId},
			errorExpected:   false,
			IdentifierPaths: []string{"$.identifierSArray[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value"},
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1N doi:10.1575/1912/bco-dmo.2343.1]",
			expectedPath:    "$.identifierSArray[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value",
			ignore:          false,
		},
		{
			name:          "identifier_obj",
			json:          map[string]string{"jsonID": jsonId},
			errorExpected: false,
			//	IdentifierPath: "$.identifierObj[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value",
			//IdentifierPath: "$.identifierObj.propertyID[@=='https://registry.identifiers.org/registry/doi')]",
			IdentifierPaths: []string{"$.identifierObj.value"},
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1]",
			expectedPath:    "$.identifierObj.value",
			ignore:          false,
		},
		//https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/earthchem2.json
		{
			name:            " identifier or id",
			json:            map[string]string{"jsonID": jsonId},
			errorExpected:   false,
			IdentifierPaths: []string{"[ $.identifiers[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value || $.['@id'] ]"},
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
		{
			name:          "identifierSArray slice",
			json:          map[string]string{"jsonID": jsonId},
			errorExpected: false,
			//IdentifierPath: "$.identifierSArray[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value[-1:]",
			IdentifierPaths: []string{"$.identifierSArray[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value.[-1:]"},
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1]",
			expectedPath:    "$.identifierSArray[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value.[-1:]",
			ignore:          true,
		},
	}

	testValidJsonPath(tests, t)
}

func TestValidJsonPathsInput(t *testing.T) {

	// this failing the first test with just one
	var jsonId = `{
"@id":"idenfitier",
"url": "http://example.com/,"
}`
	var jsonIdentifier = `{
"@id":"idenfitier",
"url": "http://example.com/",
"identifier":"doi:10"


}`
	var jsonIdentifierObject = `{
"@id":"idenfitier",
"url": "http://example.com/",
"identifier": 
	{
	"@type": "PropertyValue",
	"@id": "https://doi.org/10.1575/1912/bco-dmo.2343.1",
	"propertyID": "https://registry.identifiers.org/registry/doi",
	"value": "doi:10.1575/1912/bco-dmo.2343.1",
	"url": "https://doi.org/10.1575/1912/bco-dmo.2343.1"
	}
}`

	var jsonIdentifierArraySingle = `{
"@id":"idenfitier",
"url": "http://example.com/",
"identifier": [	
	{
	"@type": "PropertyValue",
	"@id": "https://doi.org/10.1575/1912/bco-dmo.2343.1",
	"propertyID": "https://registry.identifiers.org/registry/doi",
	"value": "doi:10.1575/1912/bco-dmo.2343.1",
	"url": "https://doi.org/10.1575/1912/bco-dmo.2343.1"
	}
]


}`
	var jsonIdentifierArrayMultiple = `{
"@id":"idenfitier",
"url": "http://example.com/",
"identifier": [	
	{
	"@type": "PropertyValue",
	"@id": "https://doi.org/10.1575/1912/bco-dmo.2343.1",
	"propertyID": "https://registry.identifiers.org/registry/doi",
	"value": "doi:10.1575/1912/bco-dmo.2343.1",
	"url": "https://doi.org/10.1575/1912/bco-dmo.2343.1"
	},
	{
	"@type": "PropertyValue",
	"@id": "https://doi.org/10.1575/1912/bco-dmo.2343.N",
	"propertyID": "https://registry.identifiers.org/registry/doi",
	"value": "doi:10.1575/1912/bco-dmo.2343.1N",
	"url": "https://doi.org/10.1575/1912/bco-dmo.2343.N"
	},
	{
	"@type": "PropertyValue",
	"@id": "https://doi.org/10.1575/1912/bco-dmo.2343.P",
	"propertyID": "https://purl.org",
	"value": "doi:10.1575/1912/bco-dmo.2343.P",
	"url": "https://doi.org/10.1575/1912/bco-dmo.2343.P"
	}
]

}`
	var tests = []expectations{
		// default
		// should work for all
		{
			name: "@id",
			json: map[string]string{"jsonID": jsonId, "jsonIdentifier": jsonIdentifier,
				"jsonobjectId":                jsonIdentifierObject,
				"jsonIdentifierArraySingle":   jsonIdentifierArraySingle,
				"jsonIdentifierArrayMultiple": jsonIdentifierArrayMultiple,
			},
			errorExpected: false,

			IdentifierPaths: []string{`$['@id']`},
			expected:        "[idenfitier]",
			expectedPath:    "$.identifierObj.value",
			ignore:          false,
		},
		//https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/earthchem2.json
		// this returns an empty set [] https://cburgmer.github.io/json-path-comparison/results/dot_notation_on_object_without_key.html
		{
			name: "$.identifier.$id",
			//json:            []string{jsonId},
			json: map[string]string{"jsonID": jsonId}, //"jsonIdentifier": jsonIdentifier,
			//"jsonobjectId": jsonIdentifierObject,
			//"jsonIdentifierArraySingle": jsonIdentifierArraySingle,
			//"jsonIdentifierArrayMultiple": jsonIdentifierArrayMultiple,

			errorExpected:   false,
			IdentifierPaths: []string{"$.identifier.value", "$.identifier", `$['@id']`},
			expected:        "[idenfitier]",
			expectedPath:    "$['@id']",
			ignore:          false,
		},
		{
			name: "$.identifier.$.identifier",
			//json:            []string{jsonIdentifier},
			json:            map[string]string{"jsonIdentifier": jsonIdentifier},
			errorExpected:   false,
			IdentifierPaths: []string{"$.identifier.value", "$.identifier", `$['@id']`},
			expected:        "[doi:10]",
			expectedPath:    "$.identifier",
			ignore:          false,
		},
		//https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/earthchem2.json
		{
			name: "$.identifierObjBracket",
			//json:            []string{jsonIdentifierObject},
			json: map[string]string{
				"jsonobjectId": jsonIdentifierObject,
			},
			errorExpected:   false,
			IdentifierPaths: []string{"$.identifier['value']", "$.identifier", `$['@id']`},
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1]",
			expectedPath:    "$.identifier['value']",
			ignore:          false,
		},
		{
			name: "$.identifierObjDot",
			//json:            []string{jsonIdentifierObject},
			json: map[string]string{
				"jsonobjectId": jsonIdentifierObject,
			},
			errorExpected:   false,
			IdentifierPaths: []string{"$.identifier.value", "$.identifier", `$['@id']`},
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1]",
			expectedPath:    "\"$.identifier.value",
			ignore:          false,
		},
		{
			name: "$.identifierObjCheck",
			//json:            []string{jsonIdentifierObject},
			json: map[string]string{
				"jsonobjectId": jsonIdentifierObject,
			},
			errorExpected:   false,
			IdentifierPaths: []string{"$.identifier.value", "$.identifier", `$['@id']`},
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1]",
			expectedPath:    "$.identifier.value",
			ignore:          false,
		},
		//https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/earthchem2.json
		{
			name: "@.identifierArraySimple",
			//json:            []string{jsonIdentifierArraySingle},
			json: map[string]string{
				"jsonIdentifierArraySingle": jsonIdentifierArraySingle,
			},
			errorExpected:   false,
			IdentifierPaths: []string{"$.identifier[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value", "$.identifier.value", "$.identifier", `$['@id']`},
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1]",
			expectedPath:    "$.identifier[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value",
			ignore:          false,
		},

		//https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/earthchem2.json
		{
			name: "@.identifierArrayMultiple",
			//json:            []string{jsonIdentifierArrayMultiple},
			json: map[string]string{
				"jsonIdentifierArrayMultiple": jsonIdentifierArrayMultiple,
			},
			errorExpected:   false,
			IdentifierPaths: []string{"$.identifier[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value", "$.identifier.value", "$.identifier", `$['@id']`},
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1N doi:10.1575/1912/bco-dmo.2343.1]",
			expectedPath:    "$.identifier[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value",
			ignore:          false,
		},
	}
	testValidJsonPaths(tests, t)
}

func testGenerateIdentifier(tests []expectations, t *testing.T) {

	var vipercontext = []byte(`
context:
  cache: true
contextmaps:
- file: ./configs/schemaorg-current-https.jsonld
  prefix: https://schema.org/
- file: ./configs/schemaorg-current-https.jsonld
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
  identifierType: filesha
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
				result, err := GenerateIdentifier(viperVal, s, json)
				//valStr := fmt.Sprint(result.uniqueId)
				assert.Equal(t, test.expected, result.uniqueId)
				assert.Equal(t, test.expectedPath, result.matchedPath)
				assert.Equal(t, test.IdentifierType, result.identifierType)
				assert.Nil(t, err)
			})
		}
	}
}

func TestGenerateIdentifier(t *testing.T) {
	var jsonIdentifierArrayMultiple = `{
"@id":"idenfitier",
"url": "http://example.com/",
"identifier": [	
	{
	"@type": "PropertyValue",
	"@id": "https://doi.org/10.1575/1912/bco-dmo.2343.1",
	"propertyID": "https://registry.identifiers.org/registry/doi",
	"value": "doi:10.1575/1912/bco-dmo.2343.1",
	"url": "https://doi.org/10.1575/1912/bco-dmo.2343.1"
	},
	{
	"@type": "PropertyValue",
	"@id": "https://doi.org/10.1575/1912/bco-dmo.2343.N",
	"propertyID": "https://registry.identifiers.org/registry/doi",
	"value": "doi:10.1575/1912/bco-dmo.2343.1N",
	"url": "https://doi.org/10.1575/1912/bco-dmo.2343.N"
	},
	{
	"@type": "PropertyValue",
	"@id": "https://doi.org/10.1575/1912/bco-dmo.2343.P",
	"propertyID": "https://purl.org",
	"value": "doi:10.1575/1912/bco-dmo.2343.P",
	"url": "https://doi.org/10.1575/1912/bco-dmo.2343.P"
	}
]

}`
	var tests = []expectations{
		// default
		// should work for all
		{
			name: "filesha",
			json: map[string]string{
				"jsonIdentifierArrayMultiple": jsonIdentifierArrayMultiple,
			},
			errorExpected:   false,
			IdentifierType:  configTypes.Filesha,
			IdentifierPaths: []string{`$['@id']`},
			expected:        "da39a3ee5e6b4b0d3255bfef95601890afd80709",
			expectedPath:    "",
			ignore:          false,
		},
		{
			name: "@id_first",
			json: map[string]string{
				"jsonIdentifierArrayMultiple": jsonIdentifierArrayMultiple,
			},
			errorExpected:   false,
			IdentifierType:  configTypes.IdentifierSha,
			IdentifierPaths: []string{`$['@id']`},
			expected:        "0fe143f05d6dbff260874a9a6e8da77243c74db0",
			expectedPath:    "$['@id']",
			ignore:          false,
		},
		{
			name: "identifier_default_path",
			json: map[string]string{
				"jsonIdentifierArrayMultiple": jsonIdentifierArrayMultiple,
			},
			errorExpected:   false,
			IdentifierType:  configTypes.IdentifierSha,
			IdentifierPaths: []string{},
			expected:        "e59f7f11a5615bcee6f35c92d8a2162e5b611944",
			expectedPath:    "$.identifier[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value",
			ignore:          false,
		},
	}

	testGenerateIdentifier(tests, t)
}
