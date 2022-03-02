package acquire

import
(
	"net/http"
	"net/http/httptest"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/spf13/viper"
)

func TestGetConfig(t *testing.T) {
	t.Run("It reads a config and returns the expected information", func(t *testing.T) {
		conf := map[string]interface{}{"minio": map[string]interface{}{"bucket": "test"}, "summoner": map[string]interface{}{"threads": "5", "delay": "0"}}

		var viper = viper.New()
		for key, value := range conf {
			viper.Set(key, value)
		}

		bucketName, tc, delay, err := getConfig(viper)
		assert.Equal(t, "test", bucketName)
		assert.Equal(t, 5, tc)
		assert.Equal(t, int64(0), delay)
		assert.Nil(t, err)
	})

	t.Run("It sets the thread count to 1 if a delay is specified", func(t *testing.T) {
		conf := map[string]interface{}{"minio": map[string]interface{}{"bucket": "test"}, "summoner": map[string]interface{}{"threads": "5", "delay": "1000"}}

		var viper = viper.New()
		for key, value := range conf {
			viper.Set(key, value)
		}

		bucketName, tc, delay, err := getConfig(viper)
		assert.Equal(t, "test", bucketName)
		assert.Equal(t, 1, tc)
		assert.Equal(t, int64(1000), delay)
		assert.Nil(t, err)
	})

	t.Run("It allows delay to be optional", func(t *testing.T) {
		conf := map[string]interface{}{"minio": map[string]interface{}{"bucket": "test"}, "summoner": map[string]interface{}{"threads": "5"}}

		var viper = viper.New()
		for key, value := range conf {
			viper.Set(key, value)
		}

		bucketName, tc, delay, err := getConfig(viper)
		assert.Equal(t, "test", bucketName)
		assert.Equal(t, 5, tc)
		assert.Equal(t, int64(0), delay)
		assert.Nil(t, err)
	})
}

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

	conf := map[string]interface{}{"sources": map[string]interface{}{"name": "test", "domain": testServer.URL }}

	var viper = viper.New()
	for key, value := range conf {
		viper.Set(key, value)
	}

	t.Run("It returns an object representing robots.txt when specified", func(t *testing.T) {
		robots, err := getRobotsForDomain(viper, "test")
		assert.NotNil(t, robots)
		assert.Nil(t, err)
	})

	t.Run("It returns nil if there is an error", func(t *testing.T) {
		robots, err := getRobotsForDomain(viper, "bad-value")
		assert.Nil(t, robots)
		assert.NotNil(t, err)
	})

}
