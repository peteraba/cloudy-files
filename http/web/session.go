package web

import (
	"fmt"
	"net/http"

	"github.com/phuslu/log"

	"github.com/peteraba/cloudy-files/apperr"
)

// FlashError sets up a temporary session message for an error.
func FlashError(w http.ResponseWriter, logger *log.Logger, err error, msg string, args ...interface{}) {
	_ = w

	problem := apperr.GetProblem(err)

	e := logger.Error().Err(err).Int("status", problem.Status)

	for i, arg := range args {
		if stringer, ok := arg.(fmt.Stringer); ok {
			e.Stringer(fmt.Sprintf("arg%d", i), stringer)
		} else {
			e.Interface(fmt.Sprintf("arg%d", i), arg)
		}
	}

	e.Msg(msg)
}

// FlashMessage sets up a temporary session message.
func FlashMessage(w http.ResponseWriter, logger *log.Logger, msg string, args ...interface{}) {
	_ = w

	e := logger.Info()

	for i, arg := range args {
		if stringer, ok := arg.(fmt.Stringer); ok {
			e.Stringer(fmt.Sprintf("arg%d", i), stringer)
		} else {
			e.Interface(fmt.Sprintf("arg%d", i), arg)
		}
	}

	e.Msg(msg)
}
