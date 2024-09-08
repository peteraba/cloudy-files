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

func setupUserHandler(t *testing.T, ctx context.Context) (http.Handler, *store.InMemory, *store.InMemory) { //nolint:unparam // sessionStore will be used soon
	t.Helper()

	factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

	userStore := store.NewInMemory(util.NewSpy())
	err := userStore.Marshal(ctx, defaultUsers)
	require.NoError(t, err)
	factory.SetStore(userStore, compose.UserStore)

	sessionStore := store.NewInMemory(util.NewSpy())
	factory.SetStore(sessionStore, compose.SessionStore)

	sut := factory.CreateUserHandler()
	handler := http.Handler(sut.SetupRoutes(http.NewServeMux()))

	return handler, userStore, sessionStore
}

func TestUserHandler_CreateUser(t *testing.T) {
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

		handler, _, _ := setupUserHandler(t, ctx)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/users", utilTest.MustReader(t, userStub))
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

		handler, _, _ := setupUserHandler(t, ctx)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/users", utilTest.MustReader(t, userStub))
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
		assert.Contains(t, actualBody, userStub.Name)
		assert.Contains(t, actualBody, userStub.Email)
	})

	t.Run("fail if request is invalid", func(t *testing.T) {
		t.Parallel()

		// setup
		handler, _, _ := setupUserHandler(t, ctx)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/users", strings.NewReader("invalid"))
		require.NoError(t, err)

		responseRecorder := httptest.NewRecorder()

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

		handler, userStoreStub, _ := setupUserHandler(t, ctx)

		userStoreStub.GetSpy().Register("ReadForWrite", 0, apperr.ErrAccessDenied)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/users", utilTest.MustReader(t, userStub))
		require.NoError(t, err)

		responseRecorder := httptest.NewRecorder()

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

func TestUserHandler_ListUsers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success json", func(t *testing.T) {
		t.Parallel()

		// setup
		handler, _, _ := setupUserHandler(t, ctx)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/users", nil)
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
		assert.Contains(t, actualBody, defaultUsers["foo"].Name)
		assert.Contains(t, actualBody, defaultUsers["bar"].Name)
	})

	t.Run("success html", func(t *testing.T) {
		t.Parallel()

		// setup
		handler, _, _ := setupUserHandler(t, ctx)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/users", nil)
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
		assert.Contains(t, actualBody, defaultUsers["foo"].Name)
		assert.Contains(t, actualBody, defaultUsers["bar"].Name)
	})

	t.Run("fail if service fails to list users", func(t *testing.T) {
		t.Parallel()

		// setup
		handler, userStoreStub, _ := setupUserHandler(t, ctx)

		userStoreSpy := userStoreStub.GetSpy()
		userStoreSpy.Register("Read", 0, apperr.ErrAccessDenied)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/users", nil)
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

func TestUserHandler_NotImplemented(t *testing.T) {
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// setup
			handler, _, _ := setupUserHandler(t, ctx)

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
			handler, _, _ := setupUserHandler(t, ctx)

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
