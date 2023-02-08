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
		name         string
		url          string
		jsonldcount  int
		headlessWait int
	}{
		{name: "r2r_1",
			url:          "https://dev.rvdata.us/search/fileset/100135",
			jsonldcount:  2,
			headlessWait: 5,
		},
	}

	for _, test := range tests {

		runstats := common.NewRepoStats(test.name)
		conf := map[string]interface{}{
			"minio":    map[string]interface{}{"bucket": "test"},
			"summoner": map[string]interface{}{"threads": "5", "delay": 10, "headless": "http://127.0.0.1:9222"},
			"sources":  []map[string]interface{}{{"name": test.name, "headlessWait": test.headlessWait}},
		}

		var viper = viper.New()
		for key, value := range conf {
			viper.Set(key, value)
		}
		repoLogger, _ := common.LogIssues(viper, test.name)
		t.Run(test.name, func(t *testing.T) {
			jsonlds, err := PageRender(viper, 5*time.Second, test.url, test.name, repoLogger, runstats)

			assert.Equal(t, test.jsonldcount, len(jsonlds))

			assert.Nil(t, err)

		})
	}
}
