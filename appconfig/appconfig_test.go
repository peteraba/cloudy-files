package appconfig_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/peteraba/cloudy-files/appconfig"
)

func TestNewConfigFromFile(t *testing.T) {
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

	t.Run("panic if file can not be loaded", func(t *testing.T) {
		t.Parallel()

		// assert
		assert.Panics(t, func() {
			_ = appconfig.NewConfigFromFile("./.envmissing")
		})
	})
}

func TestConfig_Validate(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		sut := appconfig.NewConfig()
		sut.CookieHashKey = "somethingelse"
		sut.CookieBlockKey = "somethingelse"

		sut.Validate()
	})

	t.Run("panic if cookie hash key is default", func(t *testing.T) {
		t.Parallel()

		sut := appconfig.NewConfig()
		sut.CookieBlockKey = "somethingelse"

		// assert
		assert.Panics(t, func() {
			sut.Validate()
		})
	})

	t.Run("panic if cookie block key is default", func(t *testing.T) {
		t.Parallel()

		sut := appconfig.NewConfig()
		sut.CookieHashKey = "somethingelse"

		// assert
		assert.Panics(t, func() {
			sut.Validate()
		})
	})
}
