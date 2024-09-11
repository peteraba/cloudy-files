package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peteraba/cloudy-files/appconfig"
	"github.com/peteraba/cloudy-files/apperr"
	"github.com/peteraba/cloudy-files/compose"
	composeTest "github.com/peteraba/cloudy-files/compose/test"
	"github.com/peteraba/cloudy-files/http/web"
	"github.com/peteraba/cloudy-files/repo"
	"github.com/peteraba/cloudy-files/store"
	"github.com/peteraba/cloudy-files/util"
	utilTest "github.com/peteraba/cloudy-files/util/test"
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

func setupUserHandler(t *testing.T, ctx context.Context) (http.Handler, *store.InMemory, *store.InMemory) {
	t.Helper()

	factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

	userStore := store.NewInMemory(util.NewSpy())
	err := userStore.Marshal(ctx, defaultUsers)
	require.NoError(t, err)
	factory.SetStore(userStore, compose.UserStore)

	csrfStore := store.NewInMemory(util.NewSpy())
	factory.SetStore(csrfStore, compose.CSRFStore)

	sut := factory.CreateUserHandler()
	handler := http.Handler(sut.SetupRoutes(http.NewServeMux()))

	return handler, userStore, csrfStore
}

func login(t *testing.T, r *http.Request, sessionUser repo.SessionUser) {
	t.Helper()

	factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

	cookie := factory.CreateCookieService()

	w := httptest.NewRecorder()

	cookie.StoreSessionUser(w, sessionUser)

	// r.Header["Cookie"]
	r.Header.Set("Cookie", w.Header().Get("Set-Cookie"))
}

func TestUserHandler_Login(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var (
		userStub     = defaultUsers["foo"]
		passwordStub = defaultUserPasswords["foo"]
	)

	const (
		csrfTokenStub = "f00ba7f00ba7f00ba7" //nolint:gosec // Checked
		ipAddressStub = "199.78.83.61"
	)

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// data
		formDataStub := url.Values{
			"username": {userStub.Name},
			"password": {passwordStub},
			"csrf":     {csrfTokenStub},
		}
		csrfDataStub := repo.CSRFModelMap{
			ipAddressStub: {
				{
					Token:   csrfTokenStub,
					Expires: time.Now().Add(time.Hour).Unix(),
				},
			},
		}

		// setup
		handler, _, csrfStoreStub := setupUserHandler(t, ctx)

		err := csrfStoreStub.Marshal(ctx, csrfDataStub)
		require.NoError(t, err)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/user-logins", strings.NewReader(formDataStub.Encode()))
		require.NoError(t, err)

		req.Header.Set(web.HeaderContentType, web.ContentTypeForm)
		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)
		req.RemoteAddr = ipAddressStub

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualBody := rr.Body.String()
		actualLocation := rr.Header().Get(web.HeaderLocation)

		// assert
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		assert.Equal(t, web.AfterLoginLocation, actualLocation)
		assert.Empty(t, actualBody)

		// TODO: assert flash message
	})

	t.Run("fail if parsing fails", func(t *testing.T) {
		t.Parallel()

		// setup
		handler, _, _ := setupUserHandler(t, ctx)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/user-logins", strings.NewReader("invalid"))
		require.NoError(t, err)

		req.Header.Set(web.HeaderContentType, web.ContentTypeForm)
		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)
		req.RemoteAddr = ipAddressStub

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualBody := rr.Body.String()
		actualLocation := rr.Header().Get(web.HeaderLocation)

		// assert
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		assert.NotEqual(t, web.HomeRedirectLocation, actualLocation)
		assert.Empty(t, actualBody)

		// TODO: assert flash message
	})

	t.Run("fail if csrf service fails", func(t *testing.T) {
		t.Parallel()

		// data
		formDataStub := url.Values{
			"username": {userStub.Name},
			"password": {passwordStub},
			"csrf":     {csrfTokenStub},
		}
		csrfDataStub := repo.CSRFModelMap{}

		// setup
		handler, userStoreStub, csrfStoreStub := setupUserHandler(t, ctx)

		userStoreSpy := userStoreStub.GetSpy()
		userStoreSpy.Register("Read", 0, assert.AnError)

		err := csrfStoreStub.Marshal(ctx, csrfDataStub)
		require.NoError(t, err)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/user-logins", strings.NewReader(formDataStub.Encode()))
		require.NoError(t, err)

		req.Header.Set(web.HeaderContentType, web.ContentTypeForm)
		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)
		req.RemoteAddr = ipAddressStub

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualBody := rr.Body.String()
		actualLocation := rr.Header().Get(web.HeaderLocation)

		// assert
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		assert.Equal(t, web.UserListLocation, actualLocation)
		assert.Empty(t, actualBody)

		// TODO: assert flash message
	})

	t.Run("fail if login service fails", func(t *testing.T) {
		t.Parallel()

		// data
		formDataStub := url.Values{
			"username": {userStub.Name},
			"password": {passwordStub},
			"csrf":     {csrfTokenStub},
		}
		csrfDataStub := repo.CSRFModelMap{
			ipAddressStub: {
				{
					Token:   csrfTokenStub,
					Expires: time.Now().Add(time.Hour).Unix(),
				},
			},
		}

		// setup
		handler, userStoreStub, csrfStoreStub := setupUserHandler(t, ctx)

		userStoreSpy := userStoreStub.GetSpy()
		userStoreSpy.Register("Read", 0, assert.AnError)

		err := csrfStoreStub.Marshal(ctx, csrfDataStub)
		require.NoError(t, err)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/user-logins", strings.NewReader(formDataStub.Encode()))
		require.NoError(t, err)

		req.Header.Set(web.HeaderContentType, web.ContentTypeForm)
		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)
		req.RemoteAddr = ipAddressStub

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualBody := rr.Body.String()
		actualLocation := rr.Header().Get(web.HeaderLocation)

		// assert
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		assert.Equal(t, web.UserListLocation, actualLocation)
		assert.Empty(t, actualBody)

		// TODO: assert flash message
	})
}

func TestUserHandler_CreateUser(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
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

		login(t, req, repo.SessionUser{Name: "foo", IsAdmin: true})

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualLocation := rr.Header().Get(web.HeaderLocation)
		actualBody := rr.Body.String()

		// assert
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		assert.Equal(t, web.UserListLocation, actualLocation)
		assert.Empty(t, actualBody)

		// TODO: assert flash message
	})

	t.Run("fail if no user is logged in", func(t *testing.T) {
		t.Parallel()

		// setup
		handler, _, _ := setupUserHandler(t, ctx)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/users", nil)
		require.NoError(t, err)

		responseRecorder := httptest.NewRecorder()

		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)

		// execute
		handler.ServeHTTP(responseRecorder, req)

		actualBody := responseRecorder.Body.String()
		actualLocation := responseRecorder.Header().Get(web.HeaderLocation)

		// assert
		assert.Equal(t, http.StatusSeeOther, responseRecorder.Code)
		assert.Equal(t, web.HomeRedirectLocation, actualLocation)
		assert.Empty(t, actualBody)

		// TODO: assert flash message
	})

	t.Run("fail if user is not admin", func(t *testing.T) {
		t.Parallel()

		// setup

		handler, _, _ := setupUserHandler(t, ctx)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/users", nil)
		require.NoError(t, err)

		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)
		req.Header.Set(web.HeaderContentType, web.ContentTypeForm)

		login(t, req, repo.SessionUser{Name: "foo", IsAdmin: false})

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualLocation := rr.Header().Get(web.HeaderLocation)
		actualBody := rr.Body.String()

		// assert
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		assert.Equal(t, web.AfterLoginLocation, actualLocation)
		assert.Empty(t, actualBody)

		// TODO: assert flash message
	})

	t.Run("fail if request is invalid", func(t *testing.T) {
		t.Parallel()

		// setup
		handler, _, _ := setupUserHandler(t, ctx)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/users", strings.NewReader("invalid"))
		require.NoError(t, err)

		responseRecorder := httptest.NewRecorder()

		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)

		login(t, req, repo.SessionUser{Name: "foo", IsAdmin: true})

		// execute
		handler.ServeHTTP(responseRecorder, req)

		actualBody := responseRecorder.Body.String()
		actualLocation := responseRecorder.Header().Get(web.HeaderLocation)

		// assert
		assert.Equal(t, http.StatusSeeOther, responseRecorder.Code)
		assert.Equal(t, web.UserListLocation, actualLocation)
		assert.Empty(t, actualBody)

		// TODO: assert flash message
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

		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)

		login(t, req, repo.SessionUser{Name: "foo", IsAdmin: true})

		// execute
		handler.ServeHTTP(responseRecorder, req)

		actualBody := responseRecorder.Body.String()
		actualLocation := responseRecorder.Header().Get(web.HeaderLocation)

		// assert
		assert.Equal(t, http.StatusSeeOther, responseRecorder.Code)
		assert.Equal(t, web.UserListLocation, actualLocation)
		assert.Empty(t, actualBody)

		// TODO: assert flash message
	})
}

func TestUserHandler_ListUsers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// setup
		handler, _, _ := setupUserHandler(t, ctx)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/users", nil)
		require.NoError(t, err)

		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)

		login(t, req, repo.SessionUser{Name: "foo", IsAdmin: true})

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

	t.Run("fail if user no user is logged in", func(t *testing.T) {
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
		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Contains(t, actualContentType, web.ContentTypeHTML)
		assert.Contains(t, actualBody, "</html>")
		assert.Contains(t, actualBody, "Access denied")
	})

	t.Run("fail if user is not admin", func(t *testing.T) {
		t.Parallel()

		// setup
		handler, _, _ := setupUserHandler(t, ctx)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/users", nil)
		require.NoError(t, err)

		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)

		login(t, req, repo.SessionUser{Name: "foo", IsAdmin: false})

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualBody := rr.Body.String()
		actualContentType := rr.Header().Get(web.HeaderContentType)

		// assert
		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Contains(t, actualContentType, web.ContentTypeHTML)
		assert.Contains(t, actualBody, "</html>")
		assert.Contains(t, actualBody, "Access denied")
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

		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)

		login(t, req, repo.SessionUser{Name: "foo", IsAdmin: true})

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualBody := rr.Body.String()
		actualContentType := rr.Header().Get(web.HeaderContentType)

		// assert
		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Contains(t, actualContentType, web.ContentTypeHTML)
		assert.Contains(t, actualBody, "Access denied")
	})
}

func TestUserHandler_UpdateUserPassword(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
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

		login(t, req, repo.SessionUser{Name: "foo", IsAdmin: true})

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualLocation := rr.Header().Get(web.HeaderLocation)
		actualBody := rr.Body.String()

		// assert
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		assert.Equal(t, web.UserListLocation, actualLocation)
		assert.Empty(t, actualBody)

		// TODO: assert flash message
	})

	t.Run("fail if no user is logged in", func(t *testing.T) {
		t.Parallel()

		// setup
		user := defaultUsers["bar"]

		safeURL := "/users/" + url.QueryEscape(user.Name) + "/passwords"

		handler, _, _ := setupUserHandler(t, ctx)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, safeURL, nil)
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
		assert.Equal(t, web.HomeRedirectLocation, actualLocation)
		assert.Empty(t, actualBody)

		// TODO: assert flash message
	})

	t.Run("fail if the logged in user is not an admin", func(t *testing.T) {
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

		login(t, req, repo.SessionUser{Name: "foo", IsAdmin: false})

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualLocation := rr.Header().Get(web.HeaderLocation)
		actualBody := rr.Body.String()

		// assert
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		assert.Equal(t, web.AfterLoginLocation, actualLocation)
		assert.Empty(t, actualBody)

		// TODO: assert flash message
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

		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)
		req.Header.Set(web.HeaderContentType, web.ContentTypeForm)

		login(t, req, repo.SessionUser{Name: "foo", IsAdmin: true})

		// execute
		handler.ServeHTTP(responseRecorder, req)

		actualBody := responseRecorder.Body.String()
		actualLocation := responseRecorder.Header().Get(web.HeaderLocation)

		// assert
		assert.Equal(t, http.StatusSeeOther, responseRecorder.Code)
		assert.Equal(t, web.UserListLocation, actualLocation)
		assert.Empty(t, actualBody)

		// TODO: assert flash message
	})

	t.Run("fail if service fails to update user", func(t *testing.T) {
		t.Parallel()

		// setup
		user := defaultUsers["bar"]

		formData := url.Values{
			"username": {user.Name},
			"password": {"!@iask3AI3??"},
		}

		safeURL := "/users/" + url.QueryEscape(user.Name) + "/passwords"

		handler, userStoreStub, _ := setupUserHandler(t, ctx)

		userStoreStub.GetSpy().Register("ReadForWrite", 0, apperr.ErrAccessDenied)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, safeURL, strings.NewReader(formData.Encode()))
		require.NoError(t, err)

		responseRecorder := httptest.NewRecorder()

		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)
		req.Header.Set(web.HeaderContentType, web.ContentTypeForm)

		login(t, req, repo.SessionUser{Name: "foo", IsAdmin: true})

		// execute
		handler.ServeHTTP(responseRecorder, req)

		actualLocation := responseRecorder.Header().Get(web.HeaderLocation)
		actualBody := responseRecorder.Body.String()

		// assert
		assert.Equal(t, http.StatusSeeOther, responseRecorder.Code)
		assert.Equal(t, web.UserListLocation, actualLocation)
		assert.Empty(t, actualBody)

		// TODO: assert flash message
	})
}

func TestUserHandler_UpdateUserAccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
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

		login(t, req, repo.SessionUser{Name: "foo", IsAdmin: true})

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualLocation := rr.Header().Get(web.HeaderLocation)
		actualBody := rr.Body.String()

		// assert
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		assert.Equal(t, web.UserListLocation, actualLocation)
		assert.Empty(t, actualBody)

		// TODO: assert flash message
	})

	t.Run("fail if no user is logged in", func(t *testing.T) {
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
		assert.Equal(t, web.HomeRedirectLocation, actualLocation)
		assert.Empty(t, actualBody)

		// TODO: assert flash message
	})

	t.Run("fail if the logged in user is not an admin", func(t *testing.T) {
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

		login(t, req, repo.SessionUser{Name: "foo", IsAdmin: false})

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualLocation := rr.Header().Get(web.HeaderLocation)
		actualBody := rr.Body.String()

		// assert
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		assert.Equal(t, web.AfterLoginLocation, actualLocation)
		assert.Empty(t, actualBody)

		// TODO: assert flash message
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

		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)

		login(t, req, repo.SessionUser{Name: "foo", IsAdmin: true})

		// execute
		handler.ServeHTTP(responseRecorder, req)

		actualBody := responseRecorder.Body.String()
		actualLocation := responseRecorder.Header().Get(web.HeaderLocation)

		// assert
		assert.Equal(t, http.StatusSeeOther, responseRecorder.Code)
		assert.Equal(t, web.UserListLocation, actualLocation)
		assert.Empty(t, actualBody)

		// TODO: assert flash message
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

		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)

		login(t, req, repo.SessionUser{Name: "foo", IsAdmin: true})

		// execute
		handler.ServeHTTP(responseRecorder, req)

		actualLocation := responseRecorder.Header().Get(web.HeaderLocation)
		actualBody := responseRecorder.Body.String()

		// assert
		assert.Equal(t, http.StatusSeeOther, responseRecorder.Code)
		assert.Equal(t, web.UserListLocation, actualLocation)
		assert.Empty(t, actualBody)

		// TODO: assert flash message
	})
}

func TestUserHandler_PromoteUser(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
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

		login(t, req, repo.SessionUser{Name: "foo", IsAdmin: true})

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualLocation := rr.Header().Get(web.HeaderLocation)
		actualBody := rr.Body.String()

		// assert
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		assert.Equal(t, web.UserListLocation, actualLocation)
		assert.Empty(t, actualBody)
	})

	t.Run("fail if no user is logged in", func(t *testing.T) {
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
		assert.Equal(t, web.HomeRedirectLocation, actualLocation)
		assert.Empty(t, actualBody)
	})

	t.Run("fail if the logged in user is not an admin", func(t *testing.T) {
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

		login(t, req, repo.SessionUser{Name: "foo", IsAdmin: false})

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualLocation := rr.Header().Get(web.HeaderLocation)
		actualBody := rr.Body.String()

		// assert
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		assert.Equal(t, web.AfterLoginLocation, actualLocation)
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

		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)

		login(t, req, repo.SessionUser{Name: "foo", IsAdmin: true})

		// execute
		handler.ServeHTTP(responseRecorder, req)

		actualLocation := responseRecorder.Header().Get(web.HeaderLocation)
		actualBody := responseRecorder.Body.String()

		// assert
		assert.Equal(t, http.StatusSeeOther, responseRecorder.Code)
		assert.Equal(t, web.UserListLocation, actualLocation)
		assert.Empty(t, actualBody)

		// TODO: assert flash message
	})
}

func TestUserHandler_DemoteUser(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
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

		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)

		login(t, req, repo.SessionUser{Name: "foo", IsAdmin: true})

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualBody := rr.Body.String()
		actualLocation := rr.Header().Get(web.HeaderLocation)

		// assert
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		assert.Equal(t, web.UserListLocation, actualLocation)
		assert.Empty(t, actualBody)

		// TODO: assert flash message
	})

	t.Run("fail if no user is logged in", func(t *testing.T) {
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

		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualBody := rr.Body.String()
		actualLocation := rr.Header().Get(web.HeaderLocation)

		// assert
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		assert.Equal(t, web.HomeRedirectLocation, actualLocation)
		assert.Empty(t, actualBody)

		// TODO: assert flash message
	})

	t.Run("fail if the user logged is not an admin", func(t *testing.T) {
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

		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)

		login(t, req, repo.SessionUser{Name: "foo", IsAdmin: false})

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualBody := rr.Body.String()
		actualLocation := rr.Header().Get(web.HeaderLocation)

		// assert
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		assert.Equal(t, web.AfterLoginLocation, actualLocation)
		assert.Empty(t, actualBody)

		// TODO: assert flash message
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

		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)

		login(t, req, repo.SessionUser{Name: "foo", IsAdmin: true})

		// execute
		handler.ServeHTTP(responseRecorder, req)

		actualBody := responseRecorder.Body.String()
		actualLocation := responseRecorder.Header().Get(web.HeaderLocation)

		// assert
		assert.Equal(t, http.StatusSeeOther, responseRecorder.Code)
		assert.Equal(t, web.UserListLocation, actualLocation)
		assert.Empty(t, actualBody)

		// TODO: assert flash message
	})
}

func TestUserHandler_DeleteUser(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// setup
		handler, _, _ := setupUserHandler(t, ctx)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodDelete, "/users/foo", nil)
		require.NoError(t, err)

		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)

		login(t, req, repo.SessionUser{Name: "foo", IsAdmin: true})

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualLocation := rr.Header().Get(web.HeaderLocation)
		actualBody := rr.Body.String()

		// assert
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		assert.Equal(t, web.UserListLocation, actualLocation)
		assert.Empty(t, actualBody)

		// TODO: assert flash message
	})

	t.Run("fail if no user is logged in", func(t *testing.T) {
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
		assert.Equal(t, web.HomeRedirectLocation, actualLocation)
		assert.Empty(t, actualBody)

		// TODO: assert flash message
	})

	t.Run("fail if the logged in user is not an admin", func(t *testing.T) {
		t.Parallel()

		// setup
		handler, _, _ := setupUserHandler(t, ctx)

		// setup request
		req, err := http.NewRequestWithContext(ctx, http.MethodDelete, "/users/foo", nil)
		require.NoError(t, err)

		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)

		login(t, req, repo.SessionUser{Name: "foo", IsAdmin: false})

		// execute
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		actualLocation := rr.Header().Get(web.HeaderLocation)
		actualBody := rr.Body.String()

		// assert
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		assert.Equal(t, web.AfterLoginLocation, actualLocation)
		assert.Empty(t, actualBody)

		// TODO: assert flash message
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

		req.Header.Set(web.HeaderAccept, web.ContentTypeHTML)

		login(t, req, repo.SessionUser{Name: "foo", IsAdmin: true})

		// execute
		handler.ServeHTTP(responseRecorder, req)

		actualBody := responseRecorder.Body.String()
		actualLocation := responseRecorder.Header().Get(web.HeaderLocation)

		// assert
		assert.Equal(t, http.StatusSeeOther, responseRecorder.Code)
		assert.Equal(t, web.UserListLocation, actualLocation)
		assert.Empty(t, actualBody)

		// TODO: assert flash message
	})
}
