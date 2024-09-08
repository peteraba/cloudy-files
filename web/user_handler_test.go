package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
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
		Password: "$2a$10$kOE05YXhGK5w6r9TmD7rNOLdqlcVefH9mEmXIeM4wvdlmsZCUCJMG",
		IsAdmin:  true,
		Access:   []string{"foo"},
	},
	"bar": {
		Name:     "bar",
		Email:    "bar@example.com",
		Password: "$2a$10$nfLv6lksyUkB6gApK0WYsufLtYOpRRQH4SRPRQbPQtfJoyYC.hxlS",
		IsAdmin:  false,
		Access:   []string{"bar"},
	},
}

var defaultUserPasswords = map[string]string{
	"foo": "foo1234$BarFoo",
	"bar": "bar1234$FooBar",
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

func TestUserHandler_Login(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success json", func(t *testing.T) {
		t.Parallel()

		// setup
		userStub := defaultUsers["foo"]
		passwordStub := defaultUserPasswords["foo"]
		loginStub := web.LoginRequest{
			Username: userStub.Name,
			Password: passwordStub,
		}

		handler, _, _ := setupUserHandler(t, ctx)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/user-logins", utilTest.MustReader(t, loginStub))
		require.NoError(t, err)

		req.Header.Set(web.HeaderContentType, web.ContentTypeJSON)
		req.Header.Set(web.HeaderAccept, web.ContentTypeJSON)

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualBody := rr.Body.String()
		actualContentType := rr.Header().Get(web.HeaderContentType)

		// assert
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, web.ContentTypeJSONUTF8, actualContentType)
		assert.Contains(t, actualBody, "hash")
		assert.Contains(t, actualBody, "access")
	})

	t.Run("success html", func(t *testing.T) {
		t.Parallel()

		// setup
		userStub := defaultUsers["foo"]
		passwordStub := defaultUserPasswords["foo"]
		formData := url.Values{
			"username": {userStub.Name},
			"password": {passwordStub},
		}

		handler, _, _ := setupUserHandler(t, ctx)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/user-logins", strings.NewReader(formData.Encode()))
		require.NoError(t, err)

		req.Header.Set(web.HeaderContentType, web.ContentTypeForm)
		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualBody := rr.Body.String()
		actualLocation := rr.Header().Get(web.HeaderLocation)

		// assert
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		assert.Contains(t, actualLocation, "/users")
		assert.Empty(t, actualBody)
	})

	t.Run("fail json if service fails", func(t *testing.T) {
		t.Parallel()

		// setup
		loginStub := web.LoginRequest{
			Username: "baz",
		}

		handler, _, _ := setupUserHandler(t, ctx)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/user-logins", utilTest.MustReader(t, loginStub))
		require.NoError(t, err)

		req.Header.Set(web.HeaderAccept, web.ContentTypeJSON)

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualBody := rr.Body.String()
		actualContentType := rr.Header().Get(web.HeaderContentType)

		// assert
		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Equal(t, web.ContentTypeJSONUTF8, actualContentType)
		assert.Contains(t, actualBody, "Not found")
	})

	t.Run("fail html if parsing fails", func(t *testing.T) {
		t.Parallel()

		// setup
		handler, _, _ := setupUserHandler(t, ctx)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/user-logins", strings.NewReader("invalid"))
		require.NoError(t, err)

		req.Header.Set(web.HeaderContentType, web.ContentTypeForm)
		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualBody := rr.Body.String()
		actualLocation := rr.Header().Get(web.HeaderLocation)

		// assert
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		assert.NotContains(t, actualLocation, "/users")
		assert.Empty(t, actualBody)
	})

	t.Run("fail html if service fails", func(t *testing.T) {
		t.Parallel()

		// setup
		formData := url.Values{
			"username": {"baz"},
		}

		handler, _, _ := setupUserHandler(t, ctx)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/user-logins", strings.NewReader(formData.Encode()))
		require.NoError(t, err)

		req.Header.Set(web.HeaderContentType, web.ContentTypeForm)
		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualBody := rr.Body.String()
		actualLocation := rr.Header().Get(web.HeaderLocation)

		// assert
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		assert.NotContains(t, actualLocation, "/users")
		assert.Empty(t, actualBody)
	})
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
		formData := url.Values{
			"name":     {"baz"},
			"email":    {"baz@example.com"},
			"password": {"baz1234$FooBar##!"},
		}

		handler, _, _ := setupUserHandler(t, ctx)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/users", strings.NewReader(formData.Encode()))
		require.NoError(t, err)

		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)
		req.Header.Set(web.HeaderContentType, web.ContentTypeForm)

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualLocation := rr.Header().Get(web.HeaderLocation)
		actualBody := rr.Body.String()

		// assert
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		assert.Contains(t, actualLocation, "/users")
		assert.Empty(t, actualBody)
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

func TestUserHandler_UpdateUserPassword(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success json", func(t *testing.T) {
		t.Parallel()

		// setup
		user := defaultUsers["bar"]

		data := web.PasswordChangeRequest{
			Username: user.Name,
			Password: "!@iask3AI3??",
		}

		safeURL := "/users/" + url.QueryEscape(user.Name) + "/passwords"

		handler, _, _ := setupUserHandler(t, ctx)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, safeURL, utilTest.MustReader(t, data))
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
		assert.Contains(t, actualBody, user.Name)
		assert.Contains(t, actualBody, `"password":"$`)
	})

	t.Run("success html", func(t *testing.T) {
		t.Parallel()

		// setup
		user := defaultUsers["bar"]

		formData := url.Values{
			"username": {user.Name},
			"password": {"!@iask3AI3??"},
		}

		safeURL := "/users/" + url.QueryEscape(user.Name) + "/passwords"

		handler, _, _ := setupUserHandler(t, ctx)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, safeURL, strings.NewReader(formData.Encode()))
		require.NoError(t, err)

		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)
		req.Header.Set(web.HeaderContentType, web.ContentTypeForm)

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualLocation := rr.Header().Get(web.HeaderLocation)
		actualBody := rr.Body.String()

		// assert
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		assert.Contains(t, actualLocation, "/users")
		assert.Empty(t, actualBody)
	})

	t.Run("fail if request is invalid", func(t *testing.T) {
		t.Parallel()

		// setup
		user := defaultUsers["bar"]

		safeURL := "/users/" + url.QueryEscape(user.Name) + "/passwords"
		handler, _, _ := setupUserHandler(t, ctx)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, safeURL, strings.NewReader("invalid"))
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

	t.Run("fail if service fails to update user", func(t *testing.T) {
		t.Parallel()

		// setup
		user := defaultUsers["bar"]
		require.False(t, user.IsAdmin)

		data := web.PasswordChangeRequest{
			Username: user.Name,
			Password: "!@iask3AI3??",
		}

		safeURL := "/users/" + url.QueryEscape(user.Name) + "/passwords"

		handler, userStoreStub, _ := setupUserHandler(t, ctx)

		userStoreStub.GetSpy().Register("ReadForWrite", 0, apperr.ErrAccessDenied)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, safeURL, utilTest.MustReader(t, data))
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

func TestUserHandler_UpdateUserAccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success json", func(t *testing.T) {
		t.Parallel()

		// setup
		user := defaultUsers["bar"]

		data := web.AccessChangeRequest{
			Username: user.Name,
			Access:   []string{"baz"},
		}

		safeURL := "/users/" + url.QueryEscape(user.Name) + "/accesses"

		handler, _, _ := setupUserHandler(t, ctx)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, safeURL, utilTest.MustReader(t, data))
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
		assert.Contains(t, actualBody, user.Name)
		assert.Contains(t, actualBody, `"access":["baz"]`)
	})

	t.Run("success html", func(t *testing.T) {
		t.Parallel()

		// setup
		user := defaultUsers["bar"]

		formData := url.Values{
			"username": {user.Name},
			"access":   {"baz"},
		}

		safeURL := "/users/" + url.QueryEscape(user.Name) + "/accesses"

		handler, _, _ := setupUserHandler(t, ctx)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, safeURL, strings.NewReader(formData.Encode()))
		require.NoError(t, err)

		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)
		req.Header.Set(web.HeaderContentType, web.ContentTypeForm)

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualLocation := rr.Header().Get(web.HeaderLocation)
		actualBody := rr.Body.String()

		// assert
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		assert.Contains(t, actualLocation, "/users")
		assert.Empty(t, actualBody)
	})

	t.Run("fail if request is invalid", func(t *testing.T) {
		t.Parallel()

		// setup
		user := defaultUsers["bar"]

		safeURL := "/users/" + url.QueryEscape(user.Name) + "/accesses"
		handler, _, _ := setupUserHandler(t, ctx)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, safeURL, strings.NewReader("invalid"))
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

	t.Run("fail if service fails to update user", func(t *testing.T) {
		t.Parallel()

		// setup
		user := defaultUsers["bar"]
		require.False(t, user.IsAdmin)

		userStub := web.UserNameOnlyRequest{
			Username: user.Name,
		}

		safeURL := "/users/" + url.QueryEscape(user.Name) + "/accesses"

		handler, userStoreStub, _ := setupUserHandler(t, ctx)

		userStoreStub.GetSpy().Register("ReadForWrite", 0, apperr.ErrAccessDenied)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, safeURL, utilTest.MustReader(t, userStub))
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

func TestUserHandler_PromoteUser(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success json", func(t *testing.T) {
		t.Parallel()

		// setup
		user := defaultUsers["bar"]
		require.False(t, user.IsAdmin)

		userStub := web.UserNameOnlyRequest{
			Username: user.Name,
		}

		safeURL := "/users/" + url.QueryEscape(user.Name) + "/promotions"

		handler, _, _ := setupUserHandler(t, ctx)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, safeURL, utilTest.MustReader(t, userStub))
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
		assert.Contains(t, actualBody, user.Name)
		assert.Contains(t, actualBody, `"is_admin":true`)
	})

	t.Run("success html", func(t *testing.T) {
		t.Parallel()

		// setup
		user := defaultUsers["bar"]
		require.False(t, user.IsAdmin)

		formData := url.Values{
			"username": {user.Name},
		}

		safeURL := "/users/" + url.QueryEscape(user.Name) + "/promotions"

		handler, _, _ := setupUserHandler(t, ctx)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, safeURL, strings.NewReader(formData.Encode()))
		require.NoError(t, err)

		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)
		req.Header.Set(web.HeaderContentType, web.ContentTypeForm)

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualLocation := rr.Header().Get(web.HeaderLocation)
		actualBody := rr.Body.String()

		// assert
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		assert.Contains(t, actualLocation, "/users")
		assert.Empty(t, actualBody)
	})

	t.Run("fail if service fails to update user", func(t *testing.T) {
		t.Parallel()

		// setup
		user := defaultUsers["bar"]
		require.False(t, user.IsAdmin)

		userStub := web.UserNameOnlyRequest{
			Username: user.Name,
		}

		safeURL := "/users/" + url.QueryEscape(user.Name) + "/promotions"

		handler, userStoreStub, _ := setupUserHandler(t, ctx)

		userStoreStub.GetSpy().Register("ReadForWrite", 0, apperr.ErrAccessDenied)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, safeURL, utilTest.MustReader(t, userStub))
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

func TestUserHandler_DemoteUser(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success json", func(t *testing.T) {
		t.Parallel()

		// setup
		user := defaultUsers["foo"]
		require.True(t, user.IsAdmin)

		userStub := web.UserNameOnlyRequest{
			Username: user.Name,
		}

		safeURL := "/users/" + url.QueryEscape(user.Name) + "/demotions"

		handler, _, _ := setupUserHandler(t, ctx)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, safeURL, utilTest.MustReader(t, userStub))
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
		assert.Contains(t, actualBody, user.Name)
		assert.Contains(t, actualBody, `"is_admin":false`)
	})

	t.Run("success html", func(t *testing.T) {
		t.Parallel()

		// setup
		user := defaultUsers["foo"]
		require.True(t, user.IsAdmin)

		formData := url.Values{
			"username": {user.Name},
		}

		safeURL := "/users/" + url.QueryEscape(user.Name) + "/demotions"

		handler, _, _ := setupUserHandler(t, ctx)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, safeURL, strings.NewReader(formData.Encode()))
		require.NoError(t, err)

		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)
		req.Header.Set(web.HeaderContentType, web.ContentTypeForm)

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualLocation := rr.Header().Get(web.HeaderLocation)
		actualBody := rr.Body.String()

		// assert
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		assert.Contains(t, actualLocation, "/users")
		assert.Empty(t, actualBody)
	})

	t.Run("fail if service fails to update user", func(t *testing.T) {
		t.Parallel()

		// setup
		user := defaultUsers["foo"]
		require.True(t, user.IsAdmin)

		userStub := web.UserNameOnlyRequest{
			Username: user.Name,
		}

		safeURL := "/users/" + url.QueryEscape(user.Name) + "/demotions"

		handler, userStoreStub, _ := setupUserHandler(t, ctx)

		userStoreStub.GetSpy().Register("ReadForWrite", 0, apperr.ErrAccessDenied)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, safeURL, utilTest.MustReader(t, userStub))
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

func TestUserHandler_DeleteUser(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success json", func(t *testing.T) {
		t.Parallel()

		// setup
		handler, _, _ := setupUserHandler(t, ctx)

		req, err := http.NewRequestWithContext(ctx, http.MethodDelete, "/users/foo", nil)
		require.NoError(t, err)

		req.Header.Set(web.HeaderAccept, web.ContentTypeJSON)

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualBody := rr.Body.String()

		// assert
		assert.Equal(t, http.StatusNoContent, rr.Code)
		assert.Empty(t, actualBody)
	})

	t.Run("success html", func(t *testing.T) {
		t.Parallel()

		// setup
		handler, _, _ := setupUserHandler(t, ctx)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodDelete, "/users/foo", nil)
		require.NoError(t, err)

		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualLocation := rr.Header().Get(web.HeaderLocation)
		actualBody := rr.Body.String()

		// assert
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		assert.Contains(t, actualLocation, "/users")
		assert.Empty(t, actualBody)
	})

	t.Run("fail if service fails to delete user", func(t *testing.T) {
		t.Parallel()

		// setup
		handler, userStoreStub, _ := setupUserHandler(t, ctx)

		userStoreStub.GetSpy().Register("ReadForWrite", 0, apperr.ErrAccessDenied)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodDelete, "/users/foo", nil)
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
