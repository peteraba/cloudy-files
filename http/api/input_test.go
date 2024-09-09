package api_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peteraba/cloudy-files/http/api"
)

func TestParse(t *testing.T) {
	t.Parallel()

	type foo struct {
		Name string `json:"name"`
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// setup
		req := &http.Request{
			Method: http.MethodPut,
			Body:   io.NopCloser(strings.NewReader(`{"name":"John"}`)),
		}

		// execute
		data, err := api.Parse(req, foo{})
		require.NoError(t, err)
		require.NotEmpty(t, data)

		// assert
		assert.Equal(t, "John", data.Name)
	})
}
