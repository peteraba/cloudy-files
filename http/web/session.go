package web

import (
	"fmt"
	"net/http"

	"github.com/phuslu/log"
)

// FlashError sets up logMsgWithArgs temporary session message for an error.
func FlashError(w http.ResponseWriter, logger *log.Logger, err error, msg string, args ...interface{}) {
	_ = w

	logMsgWithArgs(logger.Error().Err(err), msg, args...)
}

// FlashMessage sets up logMsgWithArgs temporary session message.
func FlashMessage(w http.ResponseWriter, logger *log.Logger, msg string, args ...interface{}) {
	_ = w

	logMsgWithArgs(logger.Info(), msg, args...)
}

func logMsgWithArgs(entry *log.Entry, msg string, args ...interface{}) {
	for i, arg := range args {
		if stringer, ok := arg.(fmt.Stringer); ok {
			entry.Stringer(fmt.Sprintf("arg%d", i), stringer)
		} else {
			entry.Interface(fmt.Sprintf("arg%d", i), arg)
		}
	}

	entry.Msg(msg)
}
