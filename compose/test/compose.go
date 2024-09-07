package test

import (
	"testing"

	"github.com/phuslu/log"

	"github.com/peteraba/cloudy-files/appconfig"
	cliTest "github.com/peteraba/cloudy-files/cli/test"
	"github.com/peteraba/cloudy-files/compose"
)

// NewTestFactory creates a new factory for test.
func NewTestFactory(t *testing.T, appConfig *appconfig.Config) *compose.Factory {
	t.Helper()

	f := compose.NewFactory(appConfig)

	f.SetLogLevel(log.PanicLevel)
	f.SetDisplay(cliTest.NewFakeDisplay(t))

	return f
}
