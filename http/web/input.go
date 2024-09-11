package web

import (
	"fmt"
	"net/http"

	"github.com/monoculum/formam/v3"

	"github.com/peteraba/cloudy-files/apperr"
	"github.com/peteraba/cloudy-files/http/inandout"
)

func GetIPAddress(r *http.Request) string {
	if r.Header.Get(inandout.HeaderXRealIP) != "" {
		return r.Header.Get(inandout.HeaderXRealIP)
	}

	if r.Header.Get(inandout.HeaderXForwardedFor) != "" {
		return r.Header.Get(inandout.HeaderXForwardedFor)
	}

	return r.RemoteAddr
}

func Parse[T any](r *http.Request, into T) (T, error) {
	err := r.ParseForm()
	if err != nil {
		return *new(T), fmt.Errorf("failed to parse form, err: %w", apperr.ErrBadRequest(err))
	}

	if len(r.Form) == 0 {
		return *new(T), fmt.Errorf("content type: %s, err: %w", r.Header.Get(inandout.HeaderContentType), apperr.ErrEmptyForm)
	}

	decoder := formam.NewDecoder(&formam.DecoderOptions{TagName: "formam"})

	err = decoder.Decode(r.Form, &into)
	if err != nil {
		return *new(T), fmt.Errorf("failed to decode %T, err: %w", into, apperr.ErrBadRequest(err))
	}

	return into, nil
}
