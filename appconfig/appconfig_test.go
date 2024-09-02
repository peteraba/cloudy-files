package appconfig_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/peteraba/cloudy-files/appconfig"
)

func TestNewConfig(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// execute
		cfg := appconfig.NewConfigFromFile("./.env")

		// assert
		assert.Equal(t, "foo", cfg.StoreAwsBucket)
		assert.Equal(t, "bar", cfg.FileSystemAwsBucket)
	})
}
