package acquire

import (
	"bytes"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

func ConfigSetupHelper(conf map[string]interface{}) *viper.Viper {
	var viper = viper.New()
	for key, value := range conf {
		viper.Set(key, value)
	}
	return viper
}

func TestGetConfig(t *testing.T) {
	t.Run("It reads a config for an indexing source and returns the expected information", func(t *testing.T) {
		conf := map[string]interface{}{
			"minio":    map[string]interface{}{"bucket": "test"},
			"summoner": map[string]interface{}{"threads": "5", "delay": 0},
			"sources":  []map[string]interface{}{{"name": "testSource"}},
		}

		viper := ConfigSetupHelper(conf)
		bucketName, tc, delay, err := getConfig(viper, "testSource")
		assert.Equal(t, "test", bucketName)
		assert.Equal(t, 5, tc)
		assert.Equal(t, int64(0), delay)
		assert.Nil(t, err)
	})

	t.Run("It sets the thread count to 1 if a delay is specified", func(t *testing.T) {
		conf := map[string]interface{}{
			"minio":    map[string]interface{}{"bucket": "test"},
			"summoner": map[string]interface{}{"threads": "5", "delay": 1000},
			"sources":  []map[string]interface{}{{"name": "testSource"}},
		}

		viper := ConfigSetupHelper(conf)
		bucketName, tc, delay, err := getConfig(viper, "testSource")
		assert.Equal(t, "test", bucketName)
		assert.Equal(t, 1, tc)
		assert.Equal(t, int64(1000), delay)
		assert.Nil(t, err)
	})

	t.Run("It allows delay to be optional", func(t *testing.T) {
		conf := map[string]interface{}{
			"minio":    map[string]interface{}{"bucket": "test"},
			"summoner": map[string]interface{}{"threads": "5"},
			"sources":  []map[string]interface{}{{"name": "testSource"}},
		}

		viper := ConfigSetupHelper(conf)
		bucketName, tc, delay, err := getConfig(viper, "testSource")
		assert.Equal(t, "test", bucketName)
		assert.Equal(t, 5, tc)
		assert.Equal(t, int64(0), delay)
		assert.Nil(t, err)
	})

	t.Run("It overrides a global summoner delay if the data source has a longer one specified", func(t *testing.T) {
		conf := map[string]interface{}{
			"minio":    map[string]interface{}{"bucket": "test"},
			"summoner": map[string]interface{}{"threads": "5", "delay": 5},
			"sources":  []map[string]interface{}{{"name": "testSource", "delay": 100}},
		}

		viper := ConfigSetupHelper(conf)
		bucketName, tc, delay, err := getConfig(viper, "testSource")
		assert.Equal(t, "test", bucketName)
		assert.Equal(t, 1, tc)
		assert.Equal(t, int64(100), delay)
		assert.Nil(t, err)
	})

	t.Run("It does not override a global summoner delay if the data source does not have a longer one specified", func(t *testing.T) {
		conf := map[string]interface{}{
			"minio":    map[string]interface{}{"bucket": "test"},
			"summoner": map[string]interface{}{"threads": "5", "delay": 50},
			"sources":  []map[string]interface{}{{"name": "testSource", "delay": 10}},
		}

		viper := ConfigSetupHelper(conf)
		bucketName, tc, delay, err := getConfig(viper, "testSource")
		assert.Equal(t, "test", bucketName)
		assert.Equal(t, 1, tc)
		assert.Equal(t, int64(50), delay)
		assert.Nil(t, err)
	})
}

func TestFindJSONInResponse(t *testing.T) {
	conf := map[string]interface{}{
		"contextmaps": map[string]interface{}{},
	}
	viper := ConfigSetupHelper(conf)
	logger := log.New()

	testJson := `{
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

	urlloc := "http://test"
	req, _ := http.NewRequest("GET", urlloc, nil)
	response := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Request:    req,
		Header:     make(http.Header, 0),
	}

	t.Run("It returns an error if the response document cannot be parsed", func(t *testing.T) {
		result, err := FindJSONInResponse(viper, urlloc, logger, nil)
		assert.Nil(t, result)
		assert.Equal(t, errors.New("Response is nil"), err)
	})

	t.Run("It finds JSON-LD in HTML document responses", func(t *testing.T) {
		html := "<html><body>yay<script type='application/ld+json'>" + testJson + "</script></body></html>"

		response.Body = ioutil.NopCloser(bytes.NewBufferString(html))
		response.ContentLength = int64(len(html))
		var expected []string

		result, err := FindJSONInResponse(viper, urlloc, logger, response)
		assert.Nil(t, err)
		assert.Equal(t, result, append(expected, testJson))
	})

	t.Run("It finds JSON-LD in JSON document responses", func(t *testing.T) {
		response.Body = ioutil.NopCloser(bytes.NewBufferString(testJson))
		response.ContentLength = int64(len(testJson))
		var expected []string

		result, err := FindJSONInResponse(viper, "test.json", logger, response)
		assert.Nil(t, err)
		assert.Equal(t, result, append(expected, testJson))
	})

	t.Run("It finds JSON-LD in http responses with a JSON-LD content type", func(t *testing.T) {
		response.Body = ioutil.NopCloser(bytes.NewBufferString(testJson))
		response.ContentLength = int64(len(testJson))
		response.Header.Add("Content-Type", JSONContentType)
		var expected []string

		result, err := FindJSONInResponse(viper, urlloc, logger, response)
		assert.Nil(t, err)
		assert.Equal(t, result, append(expected, testJson))
	})

	t.Run("It finds JSON-LD in http responses with a JSON content type", func(t *testing.T) {
		response.Body = ioutil.NopCloser(bytes.NewBufferString(testJson))
		response.ContentLength = int64(len(testJson))
		response.Header.Add("Content-Type", "application/json; charset=utf-8")
		var expected []string

		result, err := FindJSONInResponse(viper, urlloc, logger, response)
		assert.Nil(t, err)
		assert.Equal(t, result, append(expected, testJson))
	})

}
