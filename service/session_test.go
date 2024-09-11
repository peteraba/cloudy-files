package service_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peteraba/cloudy-files/appconfig"
	"github.com/peteraba/cloudy-files/apperr"
	composeTest "github.com/peteraba/cloudy-files/compose/test"
	"github.com/peteraba/cloudy-files/http/inandout"
	"github.com/peteraba/cloudy-files/repo"
	"github.com/peteraba/cloudy-files/service"
)

func TestCookie_FlashError(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// data
		expected := service.FlashMessage{Level: service.LevelError, Message: "baz"}

		// setup
		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())
		sut := factory.CreateCookieService()

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
		require.NoError(t, err)

		recorder := httptest.NewRecorder()

		// execute
		sut.FlashError(recorder, req, "/", assert.AnError, "baz", 17)

		// setup assert
		actualLocation := recorder.Header().Get(inandout.HeaderLocation)

		req.Header.Set("Cookie", recorder.Header().Get("Set-Cookie"))

		flashMessages, err := sut.GetFlashMessages(recorder, req)
		require.NoError(t, err)

		// assert
		assert.Equal(t, expected, flashMessages[0])
		assert.Equal(t, http.StatusSeeOther, recorder.Code)
		assert.Equal(t, "/", actualLocation)
	})
}

type baz int

func (b baz) String() string {
	return fmt.Sprintf("baz %d", b)
}

func TestCookie_FlashMessage(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// data
		expected := service.FlashMessage{Level: service.LevelInfo, Message: "baz"}

		// setup
		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())
		sut := factory.CreateCookieService()

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
		require.NoError(t, err)

		recorder := httptest.NewRecorder()

		// execute
		sut.FlashMessage(recorder, req, "/", "baz", baz(17))

		// setup assert
		actualLocation := recorder.Header().Get(inandout.HeaderLocation)

		req.Header.Set("Cookie", recorder.Header().Get("Set-Cookie"))

		flashMessages, err := sut.GetFlashMessages(recorder, req)
		require.NoError(t, err)

		// assert
		assert.Equal(t, expected, flashMessages[0])
		assert.Equal(t, http.StatusSeeOther, recorder.Code)
		assert.Equal(t, "/", actualLocation)
	})
}

func TestCookie_Add_and_GetFlashMessages(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// data
		expected := service.FlashMessage{Level: service.LevelInfo, Message: "baz"}

		// setup
		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())
		sut := factory.CreateCookieService()

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		// execute
		err = sut.AddFlashMessage(rr, req, expected)
		require.NoError(t, err)

		// setup assert
		req.Header.Set("Cookie", rr.Header().Get("Set-Cookie"))

		actual, err := sut.GetFlashMessages(rr, req)
		require.NoError(t, err)

		// assert
		assert.Equal(t, expected, actual[0])
	})
}

func TestCookie_GetFlashMessages(t *testing.T) {
	t.Parallel()

	t.Run("success if cookie is empty", func(t *testing.T) {
		t.Parallel()

		// setup
		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())
		sut := factory.CreateCookieService()

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
		require.NoError(t, err)

		req.Header.Set("Cookie", "flash=")

		rr := httptest.NewRecorder()

		// execute
		actual, err := sut.GetFlashMessages(rr, req)
		require.NoError(t, err)
		require.Empty(t, actual)
	})

	t.Run("fail if message can not be decoded", func(t *testing.T) {
		t.Parallel()

		// setup
		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())
		sut := factory.CreateCookieService()

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
		require.NoError(t, err)

		req.Header.Set("Cookie", "flash=bar")

		rr := httptest.NewRecorder()

		// execute
		actual, err := sut.GetFlashMessages(rr, req)
		require.Error(t, err)
		require.Empty(t, actual)

		// assert
		assert.ErrorContains(t, err, "failed to decode flash message session")
	})

	t.Run("fail if no messages", func(t *testing.T) {
		t.Parallel()

		// setup
		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())
		sut := factory.CreateCookieService()

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		// execute
		actual, err := sut.GetFlashMessages(rr, req)
		require.Error(t, err)
		require.Empty(t, actual)

		// assert
		assert.ErrorContains(t, err, "named cookie not present")
	})
}

func TestCookie_SessionUser(t *testing.T) {
	t.Parallel()

	t.Run("success store, get", func(t *testing.T) {
		t.Parallel()

		// data
		expected := repo.SessionUser{
			Name:    "baz",
			IsAdmin: true,
		}

		// setup
		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())
		sut := factory.CreateCookieService()

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
		require.NoError(t, err)

		recorder := httptest.NewRecorder()

		// execute
		sut.StoreSessionUser(recorder, expected)

		req.Header.Set("Cookie", recorder.Header().Get("Set-Cookie"))

		actual, err := sut.GetSessionUser(req)
		require.NoError(t, err)

		// assert
		assert.Equal(t, expected, actual)
	})

	t.Run("success store, get empty", func(t *testing.T) {
		t.Parallel()

		// data
		expected := repo.SessionUser{
			IsAdmin: true,
		}

		// setup
		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())
		sut := factory.CreateCookieService()

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
		require.NoError(t, err)

		recorder := httptest.NewRecorder()

		// execute
		sut.StoreSessionUser(recorder, expected)

		req.Header.Set("Cookie", recorder.Header().Get("Set-Cookie"))

		actual, err := sut.GetSessionUser(req)
		require.Error(t, err)
		require.Empty(t, actual)

		// assert
		assert.ErrorIs(t, err, apperr.ErrAccessDenied)
	})
}

func TestCookie_GetSessionUser(t *testing.T) {
	t.Parallel()

	t.Run("success get empty when not set", func(t *testing.T) {
		t.Parallel()

		// setup
		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())
		sut := factory.CreateCookieService()

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
		require.NoError(t, err)

		// execute
		sessionUser, err := sut.GetSessionUser(req)
		require.Error(t, err)
		require.Empty(t, sessionUser)

		// assert
		assert.ErrorIs(t, err, apperr.ErrAccessDenied)
	})

	t.Run("fail to get if session is missing", func(t *testing.T) {
		t.Parallel()

		// setup
		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())
		sut := factory.CreateCookieService()

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
		require.NoError(t, err)

		// execute
		sessionUser, err := sut.GetSessionUser(req)
		require.Error(t, err)
		require.Empty(t, sessionUser)

		// assert
		assert.ErrorIs(t, err, apperr.ErrAccessDenied)
	})

	t.Run("fail if message can not be decoded", func(t *testing.T) {
		t.Parallel()

		// setup
		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())
		sut := factory.CreateCookieService()

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
		require.NoError(t, err)

		req.Header.Set("Cookie", "user=bar")

		// execute
		actual, err := sut.GetSessionUser(req)
		require.Error(t, err)
		require.Empty(t, actual)

		// assert
		assert.ErrorContains(t, err, "failed to decode user session")
	})
}

func TestCookie_DeleteSessionUser(t *testing.T) {
	t.Parallel()

	t.Run("success delete when not set", func(t *testing.T) {
		t.Parallel()

		// setup
		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())
		sut := factory.CreateCookieService()

		rr := httptest.NewRecorder()

		// execute
		sut.DeleteSessionUser(rr)
	})
}
