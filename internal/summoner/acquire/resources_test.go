package acquire

import (
	configTypes "github.com/gleanerio/gleaner/internal/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/temoto/robotstxt"
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
		group := robots.FindGroup(EarthCubeAgent)
		assert.Equal(t, time.Duration(10000000000), group.CrawlDelay)
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

	robots, err := robotstxt.FromString(`User-agent: *
        Disallow: /cgi-bin
        Disallow: /forms
        Disallow: /api/gi-cat
        Disallow: /rocs/archives-catalog
        Crawl-delay: 10`)

	assert.Nil(t, err)

	group := robots.FindGroup(EarthCubeAgent)

	t.Run("It does nothing if given a nil robots object", func(t *testing.T) {
		overrideCrawlDelayFromRobots(viper, "test", 0, nil)
		sources, err := configTypes.GetSources(viper)
		source, err := configTypes.GetSourceByName(sources, "test")
		assert.Nil(t, err)
		assert.Equal(t, int64(0), source.Delay)
	})

	t.Run("It handles trying to set the crawl delay for a source that does not exist", func(t *testing.T) {
		overrideCrawlDelayFromRobots(viper, "foo", 0, group)
		assert.Nil(t, err)
	})

	t.Run("It overrides the crawl delay if it is more than our default delay", func(t *testing.T) {
		overrideCrawlDelayFromRobots(viper, "test", 9999, group)
		sources, err := configTypes.GetSources(viper)
		source, err := configTypes.GetSourceByName(sources, "test")
		assert.Nil(t, err)
		assert.Equal(t, int64(10000), source.Delay)
	})

	t.Run("It does not override the crawl delay if it is less than our default delay", func(t *testing.T) {
		overrideCrawlDelayFromRobots(viper, "test", 10001, group)
		sources, err := configTypes.GetSources(viper)
		source, err := configTypes.GetSourceByName(sources, "test")
		assert.Nil(t, err)
		assert.Equal(t, int64(10000), source.Delay)
	})
}
