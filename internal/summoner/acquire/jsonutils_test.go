package acquire

import (
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

var invalidJson = `This isn't JSON at all.`

var validJson = `{
    "@graph":[
        {
            "@context": {
                "SO":"http://schema.org/"
            },
            "@type":"bar",
            "SO:name":"Some type in a graph"
        }
    ]
}`
var arrayJsonCtx = `{
    "@graph":[
        {
            "@context": [
					"https://schema.org/",
					
				  ],
            "@type":"bar",
            "SO:name":"Some type in a graph"
        }
    ]
}`

// this has no @vocab or schema namespace defined, and is an object
var mangledJsonCtx = `{
    "@graph":[
        {
            "@context": [
					"https://schema.org/",
					{
					  "gsqtime": "https://vocabs.gsq.digital/object?uri=http://linked.data.gov.au/def/trs",
					  "time": "http://www.w3.org/2006/time#",
					  "xsd": "https://www.w3.org/TR/2004/REC-xmlschema-2-20041028/datatypes.html"
					}
				  ],
            "@type":"bar",
            "SO:name":"Some type in a graph"
        }
    ]
}`
var contextObjectGraphJson = `{
		"@context": {
		"rdf": "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
			"rdfs": "http://www.w3.org/2000/01/rdf-schema#",
			"schema": "http://schema.org/",
			"xsd": "http://www.w3.org/2001/XMLSchema#"
	    },
		"@graph": [
		      {
					"@id": "https://wifire-data.sdsc.edu/dataset/a1770ff8-1665-433c-88fb-c8e6863c61fc/resource/b01d00d2-1d64-47b8-aa5c-00410d84e6e6",
					"@type": "schema:DataDownload",
					"schema:encodingFormat": "GeoJSON",
					"schema:name": "GeoJSON",
					"schema:url": "https://gis-calema.opendata.arcgis.com/datasets/34402e97810f410db0ccd1ae345d9807_5.geojson?outSR=%7B%22latestWkid%22%3A3857%2C%22wkid%22%3A102100%7D"
				}
		]
	}
`
var v1 = viper.New()

func TestIsValid(t *testing.T) {
	t.Run("It returns true for valid JSON-LD", func(t *testing.T) {
		result, err := isValid(v1, validJson)
		assert.Equal(t, true, result)
		assert.Nil(t, err)
	})
	t.Run("It returns false and throws an error for invalid JSON-LD", func(t *testing.T) {
		result, err := isValid(v1, invalidJson)
		assert.Equal(t, false, result)
		assert.NotNil(t, err)
	})

	// to do: test for valid JSON but invalid RDF triples
}

func TestAddToJsonListIfValid(t *testing.T) {
	original := []string{"test"}

	t.Run("It appends valid json to the array", func(t *testing.T) {
		result, err := addToJsonListIfValid(v1, original, validJson)
		assert.Equal(t, []string{"test", validJson}, result)
		assert.Nil(t, err)
	})
	t.Run("It does not append invalid json to the array", func(t *testing.T) {
		result, err := addToJsonListIfValid(v1, original, invalidJson)
		assert.Equal(t, original, result)
		assert.NotNil(t, err)
	})
}

func TestContextStringFix(t *testing.T) {
	var contextObjectJson = `{
        "@context": {
            "@vocab":"http://schema.org/"
        },
        "@type":"bar",
        "SO:name":"Some type in a graph"
    }`

	var contextStringJson = `{
        "@context": "http://schema.org/",
        "@type":"bar",
        "SO:name":"Some type in a graph"
    }`

	t.Run("It rewrites the jsonld context if it is not an object", func(t *testing.T) {
		result, err := fixContextString(contextStringJson, Https)
		assert.JSONEq(t, contextObjectJson, result)
		assert.Nil(t, err)
	})

	t.Run("It does not change the jsonld context if it is already an object", func(t *testing.T) {
		result, err := fixContextString(contextObjectJson, Https)
		assert.Equal(t, contextObjectJson, result)
		assert.Nil(t, err)
	})

	t.Run("It does not change the jsonld context if it is already an object. Version 2", func(t *testing.T) {
		result, err := fixContextString(contextObjectGraphJson, Https)
		assert.Equal(t, contextObjectGraphJson, result)
		assert.Nil(t, err)
	})
}

func TestContextUrlFix(t *testing.T) {
	var httpContext = `{
"@context": {
"@vocab":"http://schema.org/"
},
"@type":"bar",
"SO:name":"Some type in a graph"
}`

	var httpNoSlashContext = `{
"@context": {
"@vocab":"http://schema.org"
},
"@type":"bar",
"SO:name":"Some type in a graph"
}`

	var noSlashContext = `{
"@context": {
"@vocab":"https://schema.org"
},
"@type":"bar",
"SO:name":"Some type in a graph"
}`

	var expectedContext = `{
"@context": {
"@vocab":"https://schema.org/"
},
"@type":"bar",
"SO:name":"Some type in a graph"
}`

	var httpContextV2 = `{
"@context": {
"@vocab":"http://schema.org/"
,"schema":"http://schema.org/"},
"@type":"bar",
"SO:name":"Some type in a graph"
}`

	var httpNoSlashContextV2 = `{
"@context": {
"@vocab":"http://schema.org"
,"schema":"http://schema.org"},
"@type":"bar",
"SO:name":"Some type in a graph"
}`

	var noSlashContextV2 = `{
"@context": {
"@vocab":"https://schema.org"
,"schema":"https://schema.org"
},
"@type":"bar",
"SO:name":"Some type in a graph"
}`

	var expectedContextv2 = `{
"@context": {
"@vocab":"https://schema.org/",
"schema":"https://schema.org/"
},
"@type":"bar",
"SO:name":"Some type in a graph"
}`
	// ContextObjectGraphJson  from global
	var expectedContextObjectGraphJson = ` {
		"@context": {
		"rdf": "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
			"rdfs": "http://www.w3.org/2000/01/rdf-schema#",
			"schema": "https://schema.org/",
			"xsd": "http://www.w3.org/2001/XMLSchema#"
	    },
		"@graph": [
		      {
					"@id": "https://wifire-data.sdsc.edu/dataset/a1770ff8-1665-433c-88fb-c8e6863c61fc/resource/b01d00d2-1d64-47b8-aa5c-00410d84e6e6",
					"@type": "schema:DataDownload",
					"schema:encodingFormat": "GeoJSON",
					"schema:name": "GeoJSON",
					"schema:url": "https://gis-calema.opendata.arcgis.com/datasets/34402e97810f410db0ccd1ae345d9807_5.geojson?outSR=%7B%22latestWkid%22%3A3857%2C%22wkid%22%3A102100%7D"
				}
		]
	}`

	t.Run("It rewrites the jsonld context if it does not have a trailing slash", func(t *testing.T) {
		result, err := fixContextUrl(noSlashContext, httpsContext)
		assert.JSONEq(t, expectedContext, result)
		assert.Nil(t, err)
	})

	t.Run("It rewrites the jsonld context if its schema is not https", func(t *testing.T) {
		result, err := fixContextUrl(httpContext, httpsContext)
		assert.JSONEq(t, expectedContext, result)
		assert.Nil(t, err)
	})

	t.Run("It rewrites the jsonld context if it does not have a trailing slash or its schema is not https", func(t *testing.T) {
		result, err := fixContextUrl(httpNoSlashContext, httpsContext)
		assert.JSONEq(t, expectedContext, result)
		assert.Nil(t, err)
	})

	//v2 fix items that are not @vocab
	t.Run("It rewrites the jsonld context if it does not have a trailing slash", func(t *testing.T) {
		result, err := fixContextUrl(noSlashContextV2, httpsContext)
		assert.JSONEq(t, expectedContextv2, result)
		assert.Nil(t, err)
	})

	t.Run("It rewrites the jsonld context if its schema is not https", func(t *testing.T) {
		result, err := fixContextUrl(httpContextV2, httpsContext)
		assert.JSONEq(t, expectedContextv2, result)
		assert.Nil(t, err)
	})

	t.Run("It rewrites the jsonld context if it does not have a trailing slash or its schema is not https", func(t *testing.T) {
		result, err := fixContextUrl(httpNoSlashContextV2, httpsContext)
		assert.JSONEq(t, expectedContextv2, result)
		assert.Nil(t, err)
	})

	t.Run("It rewrites the jsonld context so it uses https", func(t *testing.T) {
		result, err := fixContextUrl(contextObjectGraphJson, httpsContext)
		assert.JSONEq(t, expectedContextObjectGraphJson, result)
		assert.Nil(t, err)
	})

}

func TestContextArrayFix(t *testing.T) {
	var contextObjectJson = `{
	"@context": {
		"@vocab":"http://schema.org/"
	    },
     "@type":"bar",
      "SO:name":"Some type in a graph"
	}`

	var contextArrayJson = `{
        "@context": [
			{
				"@vocab": "https://schema.org/"
			},
			{
				"@vocab": "https://schema.org/",
				"NAME": "schema:name",
				"census_profile": {
				  "@id": "schema:subjectOf",
				  "@type": "@id"
			      }
			}
        ]
    }`
	var contextMixedJson = `{
        "@context": [
			
				"@vocab": "https://schema.org/"
			,
			{
				"@vocab": "https://schema.org/",
				"NAME": "schema:name",
				"census_profile": {
				"@id": "schema:subjectOf",
				"@type": "@id"
			}
			}
        ]
    }`

	// example from Magic, circa 2022-12
	// similar to spec cases 20 and 22
	//https://www.w3.org/TR/json-ld11/#example-20-describing-disconnected-nodes-with-graph
	//https://www.w3.org/TR/json-ld11/#example-22-combining-external-and-local-contexts
	//  this is a case where a string fix is needed.
	var contextLocalNamspaceJson = `{
        "@context": [
			
				 "https://schema.org/",

			{
				
				"NAME": "schema:name",
				"census_profile": {
				"@id": "schema:subjectOf",
				"@type": "@id"
			}
			}
        ]
    }`
	// is this really what we want?
	t.Run("It rewrites the jsonld context if it is not an object", func(t *testing.T) {
		result, err := fixContextArray(contextArrayJson, Https)
		assert.JSONEq(t, contextObjectJson, result)
		assert.Nil(t, err)
	})

	t.Run("It does not change the jsonld context if it is already an object", func(t *testing.T) {
		result, err := fixContextArray(contextObjectJson, Https)
		assert.Equal(t, contextObjectJson, result)
		assert.Nil(t, err)
	})

	t.Run("It does not change the jsonld context if it is mixed content", func(t *testing.T) {
		result, err := fixContextArray(contextMixedJson, Https)
		assert.Equal(t, contextMixedJson, result)
		assert.Nil(t, err)
	})
	t.Run("It should change the  the jsonld context if the 'local' namespace is a string", func(t *testing.T) {
		result, err := fixContextArray(contextLocalNamspaceJson, Https)
		assert.Equal(t, contextObjectJson, result)
		assert.Nil(t, err)
	})
}
