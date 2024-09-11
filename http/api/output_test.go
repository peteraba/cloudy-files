package api_test

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/phuslu/log"

	"github.com/peteraba/cloudy-files/http/api"
)

func TestSend(t *testing.T) {
	t.Parallel()

	t.Run("nil content", func(t *testing.T) {
		t.Parallel()

		recorder := httptest.NewRecorder()
		nomarshal := make(chan struct{})
		logger := &log.Logger{
			Level:        log.PanicLevel,
			Caller:       0,
			TimeField:    "",
			TimeFormat:   "",
			TimeLocation: nil,
			Context:      nil,
			Writer:       log.IOWriter{Writer: os.Stderr},
		}

		api.Send(recorder, nomarshal, logger)
	})
}
