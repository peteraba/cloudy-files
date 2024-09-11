package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peteraba/cloudy-files/appconfig"
	"github.com/peteraba/cloudy-files/apperr"
	"github.com/peteraba/cloudy-files/compose"
	composeTest "github.com/peteraba/cloudy-files/compose/test"
	"github.com/peteraba/cloudy-files/http/inandout"
	"github.com/peteraba/cloudy-files/repo"
	"github.com/peteraba/cloudy-files/store"
	"github.com/peteraba/cloudy-files/util"
)

func setupFallbackHandler(t *testing.T) (http.Handler, *store.InMemory) {
	t.Helper()

	factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

	csrfStoreStub := store.NewInMemory(util.NewSpy())
	factory.SetStore(csrfStoreStub, compose.CSRFStore)

	sut := factory.CreateFallbackHandler()
	handler := http.Handler(sut.SetupRoutes(http.NewServeMux()))

	return handler, csrfStoreStub
}

func TestFallbackHandler_Home(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	const (
		ipAddressStub = "132.182.38.23"
	)

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// data
		csrfDataStub := repo.CSRFModelMap{}

		// setup
		handler, csrfStoreStub := setupFallbackHandler(t)

		err := csrfStoreStub.Marshal(ctx, csrfDataStub)
		require.NoError(t, err)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/", nil)
		require.NoError(t, err)

		req.Header.Set(inandout.HeaderAccept, inandout.ContentTypeHTML)
		req.RemoteAddr = ipAddressStub

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualBody := rr.Body.String()
		actualContentType := rr.Header().Get(inandout.HeaderContentType)

		// assert
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, actualContentType, inandout.ContentTypeHTML)
		assert.Contains(t, actualBody, "</html>")
	})

	t.Run("fail if csrf can not be created", func(t *testing.T) {
		t.Parallel()

		// setup
		handler, csrfStoreStub := setupFallbackHandler(t)

		spy := csrfStoreStub.GetSpy()
		spy.Register("ReadForWrite", 0, apperr.ErrAccessDenied)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/", nil)
		require.NoError(t, err)

		req.Header.Set(inandout.HeaderAccept, inandout.ContentTypeHTML)
		req.RemoteAddr = ipAddressStub

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualBody := rr.Body.String()
		actualContentType := rr.Header().Get(inandout.HeaderContentType)

		// assert
		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Contains(t, actualContentType, inandout.ContentTypeHTML)
		assert.Contains(t, actualBody, "Access denied")
		assert.Contains(t, actualBody, "</html>")
	})
}
