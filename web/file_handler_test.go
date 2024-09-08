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
	"github.com/peteraba/cloudy-files/repo"
	"github.com/peteraba/cloudy-files/store"
	"github.com/peteraba/cloudy-files/util"
	"github.com/peteraba/cloudy-files/web"
)

func setupFileHandler(t *testing.T) (http.Handler, *store.InMemory, *store.InMemory) { //nolint:unparam // sessionStore will be used soon
	t.Helper()

	factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

	sessionStore := store.NewInMemory(util.NewSpy())
	factory.SetStore(sessionStore, compose.SessionStore)

	fileStore := store.NewInMemory(util.NewSpy())
	factory.SetStore(fileStore, compose.FileStore)

	sut := factory.CreateFileHandler()
	handler := http.Handler(sut.SetupRoutes(http.NewServeMux()))

	return handler, sessionStore, fileStore
}

func TestFileHandler_ListFiles(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success json", func(t *testing.T) {
		t.Parallel()

		// setup
		fileNameStub := "foo.txt"
		accessStub := []string{"foo", "bar"}

		filesStub := make(repo.FileModelMap, 0)
		filesStub[fileNameStub] = repo.FileModel{
			Name:   fileNameStub,
			Access: accessStub,
		}

		handler, _, fileStoreStub := setupFileHandler(t)

		err := fileStoreStub.Marshal(ctx, filesStub)
		require.NoError(t, err)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/files", nil)
		require.NoError(t, err)

		req.Header.Set(web.HeaderAccept, web.ContentTypeJSON)

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualBody := rr.Body.String()
		actualContentType := rr.Header().Get(web.HeaderContentType)

		// assert
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, actualContentType, web.ContentTypeJSON)
		assert.Contains(t, actualBody, fileNameStub)
		assert.Contains(t, actualBody, accessStub[0])
		assert.Contains(t, actualBody, accessStub[1])
	})

	t.Run("success html", func(t *testing.T) {
		t.Parallel()

		// setup
		fileNameStub := "foo.txt"
		accessStub := []string{"foo", "bar"}

		filesStub := make(repo.FileModelMap, 0)
		filesStub[fileNameStub] = repo.FileModel{
			Name:   fileNameStub,
			Access: accessStub,
		}

		handler, _, fileStoreStub := setupFileHandler(t)

		err := fileStoreStub.Marshal(ctx, filesStub)
		require.NoError(t, err)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/files", nil)
		require.NoError(t, err)

		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualBody := rr.Body.String()
		actualContentType := rr.Header().Get(web.HeaderContentType)

		// assert
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, actualContentType, web.ContentTypeHTML)
		assert.Contains(t, actualBody, "</html>")
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

		handler, _, fileStoreStub := setupFileHandler(t)

		fileStoreSpy := fileStoreStub.GetSpy()
		fileStoreSpy.Register("Read", 0, apperr.ErrAccessDenied)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/files", nil)
		require.NoError(t, err)

		req.Header.Set(web.HeaderAccept, web.ContentTypeJSON)

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualBody := rr.Body.String()
		actualContentType := rr.Header().Get(web.HeaderContentType)

		// assert
		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Contains(t, actualContentType, web.ContentTypeJSON)
		assert.Contains(t, actualBody, "Access denied")
	})
}

func TestFileHandler_NotImplemented(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name   string
		method string
		url    string
	}{
		{
			name:   "DeleteFile",
			method: http.MethodDelete,
			url:    "/files/foo",
		},
		{
			name:   "UploadFile",
			method: http.MethodPost,
			url:    "/file-uploads",
		},
		{
			name:   "RetrieveFile",
			method: http.MethodGet,
			url:    "/file-uploads",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// setup
			handler, _, _ := setupFileHandler(t)

			req, err := http.NewRequestWithContext(ctx, tt.method, tt.url, nil)
			require.NoError(t, err)

			req.Header.Set(web.HeaderAccept, web.ContentTypeJSON)

			// execute
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			actualBody := rr.Body.String()
			actualContentType := rr.Header().Get(web.HeaderContentType)

			// assert
			assert.Equal(t, http.StatusNotFound, rr.Code)
			assert.Contains(t, actualContentType, web.ContentTypeJSON)
			assert.Contains(t, actualBody, "Not Implemented.")
		})
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// setup
			handler, _, _ := setupFileHandler(t)

			req, err := http.NewRequestWithContext(ctx, tt.method, tt.url, nil)
			require.NoError(t, err)

			req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)

			// execute
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			actualBody := rr.Body.String()
			actualContentType := rr.Header().Get(web.HeaderContentType)

			// assert
			assert.Equal(t, http.StatusNotFound, rr.Code)
			assert.Contains(t, actualContentType, web.ContentTypeHTML)
			assert.Contains(t, actualBody, "</html>")
			assert.Contains(t, actualBody, "Not Implemented.")
		})
	}
}
