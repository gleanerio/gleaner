package acquire

import (
	"github.com/gleanerio/gleaner/internal/common"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test(t *testing.T) {

}

func TestHeadlessNG(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		jsonldcount int
	}{
		{name: "r2r_1",
			url:         "https://dev.rvdata.us/search/fileset/100135",
			jsonldcount: 2,
		},
	}
	conf := map[string]interface{}{
		"minio":    map[string]interface{}{"bucket": "test"},
		"summoner": map[string]interface{}{"threads": "5", "delay": 10, "headless": "http://127.0.0.1:9222"},
		"sources":  []map[string]interface{}{{"name": "testSource"}},
	}

	var viper = viper.New()
	for key, value := range conf {
		viper.Set(key, value)
	}

	for _, test := range tests {
		repoLogger, _ := common.LogIssues(viper, test.name)

		runstats := common.NewRepoStats(test.name)
		t.Run(test.name, func(t *testing.T) {
			jsonlds, err := PageRender(viper, 45*time.Second, test.url, test.name, repoLogger, runstats)

			assert.Equal(t, test.jsonldcount, len(jsonlds))

			assert.Nil(t, err)

		})
	}
}
