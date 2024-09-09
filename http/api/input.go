package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/peteraba/cloudy-files/apperr"
)

func Parse[T any](r *http.Request, into T) (T, error) {
	err := json.NewDecoder(r.Body).Decode(&into)
	if err != nil {
		return *new(T), fmt.Errorf("failed to decode %T, err: %w", into, apperr.ErrBadRequest(err))
	}

	return into, nil
}
