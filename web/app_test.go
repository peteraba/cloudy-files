package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
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
	utilTest "github.com/peteraba/cloudy-files/util/test"
	"github.com/peteraba/cloudy-files/web"
)

var defaultUsers = repo.UserModelMap{
	"foo": {
		Name:     "foo",
		Email:    "foo@example.com",
		Password: "foo1234$BarFoo",
		IsAdmin:  true,
	},
	"bar": {
		Name:     "bar",
		Email:    "bar@example.com",
		Password: "bar1234$FooBar",
		IsAdmin:  false,
	},
}

func setup(t *testing.T, ctx context.Context) (*web.App, *store.InMemory, *store.InMemory, *store.InMemory) { //nolint:unparam // sessionStore will be used soon
	t.Helper()

	factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

	userStore := store.NewInMemory(util.NewSpy())
	err := userStore.Marshal(ctx, defaultUsers)
	require.NoError(t, err)
	factory.SetStore(userStore, compose.UserStore)

	sessionStore := store.NewInMemory(util.NewSpy())
	factory.SetStore(sessionStore, compose.SessionStore)

	fileStore := store.NewInMemory(util.NewSpy())
	factory.SetStore(fileStore, compose.FileStore)

	return factory.CreateHTTPApp(), userStore, sessionStore, fileStore
}

func TestApp_CreateUser(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success json", func(t *testing.T) {
		t.Parallel()

		// setup
		userStub := repo.UserModel{
			Name:     "baz",
			Email:    "baz@example.com",
			Password: "baz1234$FooBar##!",
			IsAdmin:  false,
			Access:   []string{"baz"},
		}

		app, _, _, _ := setup(t, ctx)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/users", utilTest.MustReader(t, userStub))
		require.NoError(t, err)

		req.Header.Set(web.HeaderAccept, web.ContentTypeJSON)

		rr := httptest.NewRecorder()
		handler := http.Handler(app.Route())

		// execute
		handler.ServeHTTP(rr, req)

		actualBody := rr.Body.String()
		actualContentType := rr.Header().Get(web.HeaderContentType)

		// assert
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, actualContentType, web.ContentTypeJSON)
		assert.Contains(t, actualBody, userStub.Name)
		assert.Contains(t, actualBody, userStub.Email)
	})

	t.Run("success html", func(t *testing.T) {
		t.Parallel()

		// setup
		userStub := repo.UserModel{
			Name:     "baz",
			Email:    "baz@example.com",
			Password: "baz1234$FooBar##!",
			IsAdmin:  false,
			Access:   []string{"baz"},
		}

		app, _, _, _ := setup(t, ctx)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/users", utilTest.MustReader(t, userStub))
		require.NoError(t, err)

		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)

		rr := httptest.NewRecorder()
		handler := http.Handler(app.Route())

		// execute
		handler.ServeHTTP(rr, req)

		actualBody := rr.Body.String()
		actualContentType := rr.Header().Get(web.HeaderContentType)

		// assert
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, actualContentType, web.ContentTypeHTML)
		assert.Contains(t, actualBody, "</html>")
		assert.Contains(t, actualBody, userStub.Name)
		assert.Contains(t, actualBody, userStub.Email)
	})

	t.Run("fail if request is invalid", func(t *testing.T) {
		t.Parallel()

		// setup
		app, _, _, _ := setup(t, ctx)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/users", strings.NewReader("invalid"))
		require.NoError(t, err)

		responseRecorder := httptest.NewRecorder()
		handler := http.Handler(app.Route())

		req.Header.Set(web.HeaderAccept, web.ContentTypeJSON)

		// execute
		handler.ServeHTTP(responseRecorder, req)

		actualBody := responseRecorder.Body.String()
		actualContentType := responseRecorder.Header().Get(web.HeaderContentType)

		// assert
		assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)
		assert.Contains(t, actualContentType, web.ContentTypeJSON)
		assert.Contains(t, actualBody, "Bad request")
	})

	t.Run("fail if service fails to create user", func(t *testing.T) {
		t.Parallel()

		// setup
		userStub := repo.UserModel{
			Name:     "baz",
			Email:    "baz@example.com",
			Password: "baz1234$FooBar##!",
			IsAdmin:  false,
			Access:   []string{"baz"},
		}

		app, userStoreStub, _, _ := setup(t, ctx)

		userStoreStub.GetSpy().Register("ReadForWrite", 0, apperr.ErrAccessDenied)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/users", utilTest.MustReader(t, userStub))
		require.NoError(t, err)

		responseRecorder := httptest.NewRecorder()
		handler := http.Handler(app.Route())

		req.Header.Set(web.HeaderAccept, web.ContentTypeJSON)

		// execute
		handler.ServeHTTP(responseRecorder, req)

		actualBody := responseRecorder.Body.String()
		actualContentType := responseRecorder.Header().Get(web.HeaderContentType)

		// assert
		assert.Equal(t, http.StatusForbidden, responseRecorder.Code)
		assert.Contains(t, actualContentType, web.ContentTypeJSON)
		assert.Contains(t, actualBody, "Access denied")
	})
}

func TestApp_ListFiles(t *testing.T) {
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

		app, _, _, fileStoreStub := setup(t, ctx)

		err := fileStoreStub.Marshal(ctx, filesStub)
		require.NoError(t, err)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/files", nil)
		require.NoError(t, err)

		req.Header.Set(web.HeaderAccept, web.ContentTypeJSON)

		rr := httptest.NewRecorder()
		handler := http.Handler(app.Route())

		// execute
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

		app, _, _, fileStoreStub := setup(t, ctx)

		err := fileStoreStub.Marshal(ctx, filesStub)
		require.NoError(t, err)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/files", nil)
		require.NoError(t, err)

		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)

		rr := httptest.NewRecorder()
		handler := http.Handler(app.Route())

		// execute
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

		app, _, _, fileStoreStub := setup(t, ctx)

		fileStoreSpy := fileStoreStub.GetSpy()
		fileStoreSpy.Register("Read", 0, apperr.ErrAccessDenied)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/files", nil)
		require.NoError(t, err)

		req.Header.Set(web.HeaderAccept, web.ContentTypeJSON)

		rr := httptest.NewRecorder()
		handler := http.Handler(app.Route())

		// execute
		handler.ServeHTTP(rr, req)

		actualBody := rr.Body.String()
		actualContentType := rr.Header().Get(web.HeaderContentType)

		// assert
		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Contains(t, actualContentType, web.ContentTypeJSON)
		assert.Contains(t, actualBody, "Access denied")
	})
}

func TestApp_DeleteUser(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name   string
		method string
		url    string
	}{
		{
			name:   "GetUser",
			method: http.MethodGet,
			url:    "/users/foo",
		},
		{
			name:   "DeleteUser",
			method: http.MethodDelete,
			url:    "/users/foo",
		},
		{
			name:   "UpdateUser",
			method: http.MethodPut,
			url:    "/users/foo",
		},
		{
			name:   "DeleteFile",
			method: http.MethodDelete,
			url:    "/files/foo",
		},
		{
			name:   "DeleteUser",
			method: http.MethodDelete,
			url:    "/users/foo",
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
		{
			name:   "Home",
			method: http.MethodGet,
			url:    "/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// setup
			app, _, _, _ := setup(t, ctx)

			req, err := http.NewRequestWithContext(ctx, tt.method, tt.url, nil)
			require.NoError(t, err)

			req.Header.Set(web.HeaderAccept, web.ContentTypeJSON)

			rr := httptest.NewRecorder()
			handler := http.Handler(app.Route())

			// execute
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
			app, _, _, _ := setup(t, ctx)

			req, err := http.NewRequestWithContext(ctx, tt.method, tt.url, nil)
			require.NoError(t, err)

			req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)

			rr := httptest.NewRecorder()
			handler := http.Handler(app.Route())

			// execute
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
