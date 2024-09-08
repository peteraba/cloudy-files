package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/monoculum/formam/v3"

	"github.com/peteraba/cloudy-files/apperr"
)

func parse[T any](r *http.Request, into T) (T, error) {
	if isJSONRequest(r) {
		err := json.NewDecoder(r.Body).Decode(&into)
		if err != nil {
			return *new(T), fmt.Errorf("failed to decode %T, err: %w", into, apperr.ErrBadRequest(err))
		}

		return into, nil
	}

	err := r.ParseForm()
	if err != nil {
		return *new(T), fmt.Errorf("failed to parse form, err: %w", apperr.ErrBadRequest(err))
	}

	if len(r.Form) == 0 {
		return *new(T), fmt.Errorf("content type: %s, err: %w", r.Header.Get(HeaderContentType), apperr.ErrEmptyForm)
	}

	decoder := formam.NewDecoder(&formam.DecoderOptions{TagName: "formam"})

	err = decoder.Decode(r.Form, &into)
	if err != nil {
		return *new(T), fmt.Errorf("failed to decode %T, err: %w", into, apperr.ErrBadRequest(err))
	}

	return into, nil
}
