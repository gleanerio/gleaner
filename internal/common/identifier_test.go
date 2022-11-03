package common

import (
	"fmt"
	configTypes "github.com/gleanerio/gleaner/internal/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

type expectations struct {
	name string
	json []string

	IdentifierPaths []string
	expected        string
	errorExpected   bool `default:false`
	ignore          bool `default:false`
}

var empty = []configTypes.Sources{}

// using idenfiters as a stand in for array of identifiers.

func testValidJsonPath(tests []expectations, t *testing.T) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.ignore {
				return
			}
			result, err := GetIdentifierByPath(test.IdentifierPaths[0], test.json[0])
			valStr := fmt.Sprint(result)
			assert.Equal(t, test.expected, valStr)
			assert.Nil(t, err)
		})
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
			t.Run(fmt.Sprint(test.name, i), func(t *testing.T) {
				if test.ignore {
					return
				}
				result, err := GetIdentiferByPaths(test.IdentifierPaths, json)
				valStr := fmt.Sprint(result)
				assert.Equal(t, test.expected, valStr)
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
			json:          []string{jsonId},
			errorExpected: false,

			IdentifierPaths: []string{`$['@id']`},
			expected:        "[idenfitier]",
			ignore:          false,
		},
		//https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/earthchem2.json
		{
			name:            "@.identifier",
			json:            []string{jsonId},
			errorExpected:   false,
			IdentifierPaths: []string{"@.identifier"},
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1]",
			ignore:          false,
		},
		//https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/earthchem2.json
		{
			name:            "$.identifier",
			json:            []string{jsonId},
			errorExpected:   false,
			IdentifierPaths: []string{"$.identifier"},
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1]",
			ignore:          false,
		},
		// argo example: https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/argo.json
		{
			name:            "identifiers Array ",
			json:            []string{jsonId},
			errorExpected:   false,
			IdentifierPaths: []string{"$.identifierSArray[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value"},
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1N doi:10.1575/1912/bco-dmo.2343.1]",
			ignore:          false,
		},
		{
			name:          "identifier_obj",
			json:          []string{jsonId},
			errorExpected: false,
			//	IdentifierPath: "$.identifierObj[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value",
			//IdentifierPath: "$.identifierObj.propertyID[@=='https://registry.identifiers.org/registry/doi')]",
			IdentifierPaths: []string{"$.identifierObj.value"},
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1]",
			ignore:          false,
		},
		//https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/earthchem2.json
		{
			name:            " identifier or id",
			json:            []string{jsonId},
			errorExpected:   false,
			IdentifierPaths: []string{"[ $.identifiers[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value || $.['@id'] ]"},
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1]",
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
			json:          []string{jsonId},
			errorExpected: false,
			//IdentifierPath: "$.identifierSArray[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value[-1:]",
			IdentifierPaths: []string{"$.identifierSArray[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value.[-1:]"},
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1]",
			ignore:          true,
		},
	}

	testValidJsonPath(tests, t)
}

func TestValidJsonPathsInput(t *testing.T) {

	// this failing the first test with just one
	var jsonId = `{
"@id":"idenfitier",
"any": "any"
}`
	var jsonIdentifier = `{
"@id":"idenfitier",
"identifier":"doi:10.1575/1912/bco-dmo.2343.1"


}`
	var jsonIdentifierObject = `{
"@id":"idenfitier",
"idenfitier": 
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
			name:          "@id",
			json:          []string{jsonId, jsonIdentifier, jsonIdentifierObject, jsonIdentifierArraySingle, jsonIdentifierArrayMultiple},
			errorExpected: false,

			IdentifierPaths: []string{`$['@id']`},
			expected:        "[idenfitier]",
			ignore:          false,
		},
		//https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/earthchem2.json
		{
			name:            "$.identifier.$id",
			json:            []string{jsonId},
			errorExpected:   false,
			IdentifierPaths: []string{"$.identifier.value", "$.identifier", `$['@id']`},
			expected:        "[idenfitier]",
			ignore:          false,
		},

		//https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/earthchem2.json
		{
			name:            "$.identifierObj",
			json:            []string{jsonIdentifier, jsonIdentifierObject},
			errorExpected:   false,
			IdentifierPaths: []string{"$.identifier.value", "$.identifier", `$['@id']`},
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1]",
			ignore:          false,
		},

		//https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/earthchem2.json
		{
			name:            "@.identifierArraySimple",
			json:            []string{jsonIdentifierArraySingle},
			errorExpected:   false,
			IdentifierPaths: []string{"$.identifier[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value", "$.identifier.value", "$.identifier", `$['@id']`},
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1]",
			ignore:          false,
		},

		//https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/earthchem2.json
		{
			name:            "@.identifierArrayMultiple",
			json:            []string{jsonIdentifierArrayMultiple},
			errorExpected:   false,
			IdentifierPaths: []string{"$.identifier[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value", "$.identifier.value", "$.identifier", `$['@id']`},
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1N doi:10.1575/1912/bco-dmo.2343.1]",
			ignore:          false,
		},
	}
	testValidJsonPaths(tests, t)
}
