package web_test

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/peteraba/cloudy-files/appconfig"
	composeTest "github.com/peteraba/cloudy-files/compose/test"
	"github.com/peteraba/cloudy-files/http"
)

// This test lives in the web package, because moving it to the http package would mess up the code coverage.
func TestApp_Start(t *testing.T) {
	t.Parallel()

	setup := func(t *testing.T) *http.App {
		t.Helper()

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

		return factory.CreateHTTPApp()
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// skip this test on OSX as it's annoying to approve the listener every time
		if runtime.GOOS == "darwin" {
			t.SkipNow()
		}

		// setup
		sut := setup(t)

		// execute
		mux := sut.Route()

		go func() {
			sut.Start(mux)
		}()

		time.Sleep(1 * time.Second)

		signal.Notify(make(chan os.Signal, 1), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

		// assert
		require.NotNil(t, mux)
	})
}
