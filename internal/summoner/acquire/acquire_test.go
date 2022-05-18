package acquire

import (
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGetConfig(t *testing.T) {
	t.Run("It reads a config for an indexing source and returns the expected information", func(t *testing.T) {
		conf := map[string]interface{}{
			"minio":    map[string]interface{}{"bucket": "test"},
			"summoner": map[string]interface{}{"threads": "5", "delay": 0},
			"sources":  []map[string]interface{}{{"name": "testSource"}},
		}

		var viper = viper.New()
		for key, value := range conf {
			viper.Set(key, value)
		}

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

		var viper = viper.New()
		for key, value := range conf {
			viper.Set(key, value)
		}

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

		var viper = viper.New()
		for key, value := range conf {
			viper.Set(key, value)
		}

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

		var viper = viper.New()
		for key, value := range conf {
			viper.Set(key, value)
		}

		bucketName, tc, delay, err := getConfig(viper, "testSource")
		assert.Equal(t, "test", bucketName)
		assert.Equal(t, 1, tc)
		assert.Equal(t, int64(100), delay)
		assert.Nil(t, err)
	})

	t.Run("It does not override a global summoner delay if the data source does not have a longer one specified", func(t *testing.T) {
		conf := map[string]interface{}{"minio": map[string]interface{}{"bucket": "test"}, "summoner": map[string]interface{}{"threads": "5", "delay": 50}, "sources": []map[string]interface{}{{"name": "testSource", "delay": 10}}}

		var viper = viper.New()
		for key, value := range conf {
			viper.Set(key, value)
		}

		bucketName, tc, delay, err := getConfig(viper, "testSource")
		assert.Equal(t, "test", bucketName)
		assert.Equal(t, 1, tc)
		assert.Equal(t, int64(50), delay)
		assert.Nil(t, err)
	})
}
