package acquire

import (
	"testing"
    "github.com/stretchr/testify/assert"
    configTypes "github.com/gleanerio/gleaner/internal/config"
)

func TestRetrieveAPIEndpoints(t *testing.T) {
	t.Run("It reads a config for an API indexing source and returns the expected information", func(t *testing.T) {
        apiSource := configTypes.Sources{
            Name: "apiSource",
            SourceType: "api",
            Active: true,
        }
        conf := map[string]interface{}{
            "sources":  []map[string]interface{}{
                {
                    "name": "testSource",
                    "sourcetype": "test",
                    "active": "true",

                },
                {
                    "name": "sitemapSource",
                    "sourcetype": "sitemap",
                    "active": "true",
                },
                {
                    "name": "apiSource",
                    "sourcetype": "api",
                    "active": "true",
                },
            },
        }

        viper := ConfigSetupHelper(conf)
        sources, err := RetrieveAPIEndpoints(viper)
        var expected []Sources
        assert.Equal(t, append(expected, apiSource), sources)
        assert.Nil(t, err)
	})
}
