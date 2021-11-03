package acquire

import
(
	"testing"
	"github.com/stretchr/testify/assert"
    "github.com/spf13/viper"
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
        assert.Equal(t, result, true)
        assert.Nil(t, err)
    })
    t.Run("It returns false and throws an error for invalid JSON-LD", func(t *testing.T) {
        result, err := isValid(v1, invalidJson)
        assert.Equal(t, result, false)
        assert.NotNil(t, err)
    })

    // to do: test for valid JSON but invalid RDF triples
}

func TestAddToJsonListIfValid(t *testing.T) {
    original := []string{"test"}

    t.Run("It appends valid json to the array", func(t *testing.T) {
        result, err := addToJsonListIfValid(v1, original, validJson)
        assert.Equal(t, result, []string{"test", validJson})
        assert.Nil(t, err)
    })
    t.Run("It does not append invalid json to the array", func(t *testing.T) {
        result, err := addToJsonListIfValid(v1, original, invalidJson)
        assert.Equal(t, result, original)
        assert.NotNil(t, err)
    })
}

// TODO: test Upload by mocking out a bunch of stuff like MinioClient
