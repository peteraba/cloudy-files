package api_test

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

func setupFileHandler(t *testing.T) (http.Handler, *store.InMemory) {
	t.Helper()

	factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

	fileStore := store.NewInMemory(util.NewSpy())
	factory.SetStore(fileStore, compose.FileStore)

	sut := factory.CreateFileHandler()
	handler := http.Handler(sut.SetupRoutes(http.NewServeMux()))

	return handler, fileStore
}

func TestFileHandler_ListFiles(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// setup
		fileNameStub := "foo.txt"
		accessStub := []string{"foo", "bar"}

		filesStub := make(repo.FileModelMap, 0)
		filesStub[fileNameStub] = repo.FileModel{
			Name:   fileNameStub,
			Access: accessStub,
		}

		handler, fileStoreStub := setupFileHandler(t)

		err := fileStoreStub.Marshal(ctx, filesStub)
		require.NoError(t, err)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/files", nil)
		require.NoError(t, err)

		req.Header.Set(inandout.HeaderAccept, inandout.ContentTypeJSON)

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualBody := rr.Body.String()
		actualContentType := rr.Header().Get(inandout.HeaderContentType)

		// assert
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, actualContentType, inandout.ContentTypeJSON)
		assert.Contains(t, actualBody, fileNameStub)
		assert.Contains(t, actualBody, accessStub[0])
		assert.Contains(t, actualBody, accessStub[1])
	})

	t.Run("fail if service fails to list files", func(t *testing.T) {
		t.Parallel()

		// setup
		fileNameStub := "foo.txt"
		accessStub := []string{"foo", "bar"}

		filesStub := make(repo.FileModelMap, 0)
		filesStub[fileNameStub] = repo.FileModel{
			Name:   fileNameStub,
			Access: accessStub,
		}

		handler, fileStoreStub := setupFileHandler(t)

		fileStoreSpy := fileStoreStub.GetSpy()
		fileStoreSpy.Register("Read", 0, apperr.ErrAccessDenied)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/files", nil)
		require.NoError(t, err)

		req.Header.Set(inandout.HeaderAccept, inandout.ContentTypeJSON)

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualBody := rr.Body.String()
		actualContentType := rr.Header().Get(inandout.HeaderContentType)

		// assert
		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Contains(t, actualContentType, inandout.ContentTypeJSON)
		assert.Contains(t, actualBody, "Access denied")
	})
}

func TestFileHandler_NotImplemented(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("delete", func(t *testing.T) {
		t.Parallel()

		// setup
		handler, _ := setupFileHandler(t)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodDelete, "/files/foo", nil)
		require.NoError(t, err)

		req.Header.Set(inandout.HeaderAccept, inandout.ContentTypeJSON)

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualBody := rr.Body.String()
		actualContentType := rr.Header().Get(inandout.HeaderContentType)

		// assert
		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, actualContentType, inandout.ContentTypeJSON)
		assert.Contains(t, actualBody, "Not implemented")
	})
}
