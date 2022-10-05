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
		result, err := fixContextString(contextStringJson)
		assert.JSONEq(t, contextObjectJson, result)
		assert.Nil(t, err)
	})

	t.Run("It does not change the jsonld context if it is already an object", func(t *testing.T) {
		result, err := fixContextString(contextObjectJson)
		assert.Equal(t, contextObjectJson, result)
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

	t.Run("It rewrites the jsonld context if it does not have a trailing slash", func(t *testing.T) {
		result, err := fixContextUrl(noSlashContext)
		assert.JSONEq(t, expectedContext, result)
		assert.Nil(t, err)
	})

	t.Run("It rewrites the jsonld context if its schema is not https", func(t *testing.T) {
		result, err := fixContextUrl(httpContext)
		assert.JSONEq(t, expectedContext, result)
		assert.Nil(t, err)
	})

	t.Run("It rewrites the jsonld context if it does not have a trailing slash or its schema is not https", func(t *testing.T) {
		result, err := fixContextUrl(httpNoSlashContext)
		assert.JSONEq(t, expectedContext, result)
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

	t.Run("It rewrites the jsonld context if it is not an object", func(t *testing.T) {
		result, err := fixContextString(contextArrayJson)
		assert.JSONEq(t, contextObjectJson, result)
		assert.Nil(t, err)
	})

	t.Run("It does not change the jsonld context if it is already an object", func(t *testing.T) {
		result, err := fixContextString(contextObjectJson)
		assert.Equal(t, contextObjectJson, result)
		assert.Nil(t, err)
	})
}
