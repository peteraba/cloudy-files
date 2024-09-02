package appconfig_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/peteraba/cloudy-files/appconfig"
)

func TestNewConfig(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		_, err := os.Stat("./.env")
		if os.IsNotExist(err) {
			t.SkipNow()
		}

		// execute
		cfg := appconfig.NewConfigFromFile("./.env")

		// assert
		assert.Equal(t, "foo", cfg.StoreAwsBucket)
		assert.Equal(t, "bar", cfg.FileSystemAwsBucket)
	})
}
