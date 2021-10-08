package acquire

import
(
	"fmt"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

type findTypeTest struct {
	inputJSON string
	expected string
	err error
}

var jsonTopLevel = `{
    "@context":"http://schema.org/",
    "@type":"foo",
    "name":"http remote @context"
}`

var jsonGraph = `{
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

var jsonInvalidType = `{
    "@context":"http://schema.org/",
    "@type":42,
    "name":"http remote @context"
}`

var jsonNoGraph = `{
    "@context":"http://schema.org/",
    "@foo":"bar",
    "name":"http remote @context"
}`

var jsonNoType = `{
    "@graph":[
        {
            "@context": {
                "SO":"http://schema.org/"
            },
            "@foo":"bar",
            "SO:name":"Some foo in a graph"
        }
    ]
}`

var findTypeTests =[]findTypeTest {
	findTypeTest{jsonTopLevel, "foo", nil},
	findTypeTest{jsonGraph, "bar", nil},
	findTypeTest{jsonInvalidType, "", fmt.Errorf("Found invalid json object for type")},
	findTypeTest{jsonNoType, "", fmt.Errorf("No type or graph found. Exiting.")},
	findTypeTest{jsonNoGraph, "", fmt.Errorf("No type or graph found. Exiting.")},
}

func TestFindType(t* testing.T) {
    for _, test := range findTypeTests {
    	var jsonInterface map[string]interface{}
    	json.Unmarshal([]byte(test.inputJSON), &jsonInterface)
        output, err := findType(jsonInterface)
        assert.Equal(t, output, test.expected)
		assert.Equal(t, err, test.err)
	}

}
