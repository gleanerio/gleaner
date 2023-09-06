package acquire

import (
	"github.com/gleanerio/gleaner/internal/common"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

var HEADLESS_URL = "http://127.0.0.1:9222"

func PingHeadless() (int, error) {
	var client = http.Client{
		Timeout: 2 * time.Second,
	}

	req, err := http.NewRequest("HEAD", HEADLESS_URL, nil)
	if err != nil {
		return 0, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	resp.Body.Close()
	return resp.StatusCode, nil
}

func TestHeadlessNG(t *testing.T) {
	status, err := PingHeadless()

	if err != nil || status != 200 {
		t.Skip("Skipping headless tests because no headless browser is running.")
	}

	tests := []struct {
		name         string
		url          string
		jsonldcount  int
		headlessWait int
		expectedFail bool "default:false"
	}{
		{name: "r2r_wait_5_works_returns_2_jsonld",
			url:          "https://dev.rvdata.us/search/fileset/100135",
			jsonldcount:  2,
			headlessWait: 5,
		},
		{name: "r2r_expectedfail_wait_0_returns_1_jsonld_fails_if_2_jsonld",
			url:          "https://dev.rvdata.us/search/fileset/100135",
			jsonldcount:  2,
			headlessWait: 0,
			expectedFail: true,
		},
	}

	for _, test := range tests {

		runstats := common.NewRepoStats(test.name)
		conf := map[string]interface{}{
			"minio":    map[string]interface{}{"bucket": "test"},
			"summoner": map[string]interface{}{"threads": "5", "delay": 10, "headless": HEADLESS_URL},
			"sources":  []map[string]interface{}{{"name": test.name, "headlessWait": test.headlessWait}},
		}

		var viper = viper.New()
		for key, value := range conf {
			viper.Set(key, value)
		}
		repoLogger, _ := common.LogIssues(viper, test.name)
		t.Run(test.name, func(t *testing.T) {
			jsonlds, err := PageRender(viper, 5*time.Second, test.url, test.name, repoLogger, runstats, nil)
			if !test.expectedFail {
				assert.Equal(t, test.jsonldcount, len(jsonlds))
			} else {
				assert.NotEqual(t, test.jsonldcount, len(jsonlds))
			}

			assert.Nil(t, err)

		})
	}
}
