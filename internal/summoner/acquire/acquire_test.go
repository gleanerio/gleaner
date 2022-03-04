package acquire

import
(
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
