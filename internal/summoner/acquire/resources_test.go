package acquire

import (
	"github.com/samclarke/robotstxt"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetRobotsForDomain(t *testing.T) {
	var robots = `User-agent: *
        Disallow: /cgi-bin
        Disallow: /forms
        Disallow: /api/gi-cat
        Disallow: /rocs/archives-catalog
        Crawl-delay: 10`

	mux := http.NewServeMux()

	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte(robots))
	})
	// generate a test server so we can capture and inspect the request
	testServer := httptest.NewServer(mux)
	defer func() { testServer.Close() }()

	t.Run("It returns an object representing robots.txt when specified", func(t *testing.T) {
		robots, err := getRobotsForDomain(testServer.URL)
		assert.NotNil(t, robots)
		assert.Nil(t, err)
		assert.Equal(t, time.Duration(10000000000), robots.CrawlDelay("*"))
	})

	t.Run("It returns nil if there is an error", func(t *testing.T) {
		robots, err := getRobotsForDomain(testServer.URL + "/bad-value")
		assert.Nil(t, robots)
		assert.NotNil(t, err)
	})
}

func TestOverrideCrawlDelayFromRobots(t *testing.T) {
	conf := map[string]interface{}{
		"sources": []map[string]string{
			{"name": "test", "domain": "http://test.com"},
		},
	}

	var viper = viper.New()
	for key, value := range conf {
		viper.Set(key, value)
	}

	robots, err := robotstxt.Parse(`User-agent: *
        Disallow: /cgi-bin
        Disallow: /forms
        Disallow: /api/gi-cat
        Disallow: /rocs/archives-catalog
        Crawl-delay: 10`, "http://www.test.com/robots.txt")

    assert.Nil(t, err)

	t.Run("It does nothing if given a nil robots object", func(t *testing.T) {
		overrideCrawlDelayFromRobots(v1, "test", 0, nil)
		assert.Nil(t, v1.Get("sources.test.delay"))
	})

	t.Run("It handles trying to set the crawl delay for a source that does not exist", func(t *testing.T) {
		overrideCrawlDelayFromRobots(v1, "foo", 0, robots)
		assert.Nil(t, v1.Get("sources.test.delay"))
		assert.Nil(t, v1.Get("sources.foo.delay"))

	})
}
