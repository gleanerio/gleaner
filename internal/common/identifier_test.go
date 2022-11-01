package common

import (
	"fmt"
	configTypes "github.com/gleanerio/gleaner/internal/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

var sources = []configTypes.Sources{
	// default
	{
		Name:           "test1",
		Headless:       true,
		Active:         true,
		SourceType:     "sitemap",
		IdentifierPath: `$["@id"]`,
	},
	//https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/earthchem2.json
	{
		Name:           "test3",
		Headless:       false,
		Active:         false,
		SourceType:     "sitemap",
		IdentifierPath: "@.identifier",
	},
	//https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/earthchem2.json
	{
		Name:           "test3",
		Headless:       false,
		Active:         false,
		SourceType:     "sitemap",
		IdentifierPath: "$.identifier",
	},
	// argo example: https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/argo.json
	{
		Name:           "test2",
		Headless:       false,
		Active:         true,
		SourceType:     "sitemap",
		IdentifierPath: "$.identifiers[?(@propertyId='https://registry.identifiers.org/registry/doi')].value",
	},

	//https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/earthchem2.json
	{
		Name:           "test4",
		Headless:       false,
		Active:         false,
		SourceType:     "sitemap",
		IdentifierPath: "[ $.identifiers[?(@propertyId='https://registry.identifiers.org/registry/doi')].value, $['@id'] ]",
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
		Name:           "test4",
		Headless:       true,
		Active:         false,
		SourceType:     "sitemap",
		IdentifierPath: "$.identifiers[?(@propertyId='https://registry.identifiers.org/registry/doi')].value[-1:]",
	},
}

var empty = []configTypes.Sources{}

// using idenfiters as a stand in for array of identifiers.
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
"identifier_obj": 
	{
	"@type": "PropertyValue",
	"@id": "https://doi.org/10.1575/1912/bco-dmo.2343.1",
	"propertyID": "https://registry.identifiers.org/registry/doi",
	"value": "doi:10.1575/1912/bco-dmo.2343.1",
	"url": "https://doi.org/10.1575/1912/bco-dmo.2343.1"
	}

}`

func TestIsValid(t *testing.T) {

	t.Run("@id", func(t *testing.T) {

		result, err := GetIdentifierByPath(sources[0].IdentifierPath, jsonId)
		valStr := fmt.Sprint(result)
		assert.Equal(t, "[idenfitier]", valStr)
		assert.Nil(t, err)
	})
	t.Run(".idenfitier", func(t *testing.T) {

		result, err := GetIdentifierByPath(sources[1].IdentifierPath, jsonId)
		valStr := fmt.Sprint(result)
		assert.Equal(t, "[doi:10.1575/1912/bco-dmo.2343.1]", valStr)
		assert.Nil(t, err)
	})
	t.Run("$.idenfitier", func(t *testing.T) {

		result, err := GetIdentifierByPath(sources[2].IdentifierPath, jsonId)
		valStr := fmt.Sprint(result)
		assert.Equal(t, "[doi:10.1575/1912/bco-dmo.2343.1]", valStr)
		assert.Nil(t, err)
	})
	// to do: test for valid JSON but invalid RDF triples
}
