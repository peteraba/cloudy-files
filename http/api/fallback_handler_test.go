package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peteraba/cloudy-files/appconfig"
	composeTest "github.com/peteraba/cloudy-files/compose/test"
	"github.com/peteraba/cloudy-files/http/api"
)

func setupFallbackHandler(t *testing.T) http.Handler {
	t.Helper()

	factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

	sut := factory.CreateFallbackHandler()
	handler := http.Handler(sut.SetupRoutes(http.NewServeMux()))

	return handler
}

func TestFallbackHandler_Home(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success json", func(t *testing.T) {
		t.Parallel()

		// setup
		handler := setupFallbackHandler(t)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/", nil)
		require.NoError(t, err)

		req.Header.Set(api.HeaderAccept, api.ContentTypeJSON)

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualBody := rr.Body.String()
		actualContentType := rr.Header().Get(api.HeaderContentType)

		// assert
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, actualContentType, api.ContentTypeJSON)
		assert.JSONEq(t, `{"status":"ok"}`, actualBody)
	})
}
