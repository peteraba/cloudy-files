package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peteraba/cloudy-files/appconfig"
	"github.com/peteraba/cloudy-files/apperr"
	"github.com/peteraba/cloudy-files/compose"
	composeTest "github.com/peteraba/cloudy-files/compose/test"
	"github.com/peteraba/cloudy-files/repo"
	"github.com/peteraba/cloudy-files/service"
	"github.com/peteraba/cloudy-files/store"
	"github.com/peteraba/cloudy-files/util"
)

func TestSession_Check(t *testing.T) {
	t.Parallel()

	unusedSpy := util.NewSpy()
	ctx := context.Background()

	setup := func(t *testing.T, sessionStoreSpy, userStoreSpy *util.Spy) (*service.Session, *service.User) {
		t.Helper()

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

		factory.SetStore(store.NewInMemory(sessionStoreSpy), compose.SessionStore)
		factory.SetStore(store.NewInMemory(userStoreSpy), compose.UserStore)

		return factory.CreateSessionService(), factory.CreateUserService()
	}

	t.Run("fails when store returns error", func(t *testing.T) {
		t.Parallel()

		// data
		stubPassword := gofakeit.Password(true, true, true, true, false, 16)
		stubHash := gofakeit.UUID()

		// setup
		sessionStoreSpy := util.NewSpy()
		sessionStoreSpy.Register("Read", 0, assert.AnError)

		sut, _ := setup(t, sessionStoreSpy, unusedSpy)

		// execute
		err := sut.Check(ctx, stubPassword, stubHash)
		require.Error(t, err)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("fail when session is checked without login", func(t *testing.T) {
		t.Parallel()

		// data
		stubName := gofakeit.Name()
		stubEmail := gofakeit.Email()
		stubPassword := gofakeit.Password(true, true, true, true, false, 16)
		stubAccess := []string{gofakeit.Adverb(), gofakeit.Adverb()}
		stubHash := gofakeit.UUID()

		// setup
		sut, userService := setup(t, unusedSpy, unusedSpy)

		userModel, err := userService.Create(ctx, stubName, stubEmail, stubPassword, false, stubAccess)
		require.NoError(t, err)
		require.NotEmpty(t, userModel)

		// execute
		err = sut.Check(ctx, stubName, stubHash)
		require.Error(t, err)

		// assert
		assert.ErrorIs(t, err, apperr.ErrNotFound)
	})

	t.Run("logged in user has valid session", func(t *testing.T) {
		t.Parallel()

		// data
		stubName := gofakeit.Name()
		stubEmail := gofakeit.Email()
		stubPassword := gofakeit.Password(true, true, true, true, false, 16)
		stubAccess := []string{gofakeit.Adverb(), gofakeit.Adverb()}

		// setup
		sut, userService := setup(t, unusedSpy, unusedSpy)

		userModel, err := userService.Create(ctx, stubName, stubEmail, stubPassword, false, stubAccess)
		require.NoError(t, err)
		require.NotEmpty(t, userModel)

		sessionModel, err := userService.Login(ctx, stubName, stubPassword)
		require.NoError(t, err)
		require.NotEmpty(t, sessionModel)

		// execute
		err = sut.Check(ctx, stubName, sessionModel.Hash)

		// assert
		require.NoError(t, err)
	})
}

func TestSession_CleanUp(t *testing.T) {
	t.Parallel()

	unusedSpy := util.NewSpy()
	ctx := context.Background()

	setup := func(t *testing.T, sessionStoreSpy *util.Spy, sessionData repo.SessionModelMap) *service.Session {
		t.Helper()

		sessionStore := store.NewInMemory(sessionStoreSpy)
		err := sessionStore.Marshal(ctx, sessionData)
		require.NoError(t, err)

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

		factory.SetStore(sessionStore, compose.SessionStore)

		return factory.CreateSessionService()
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// setup
		sut := setup(t, unusedSpy, repo.SessionModelMap{"peter": {Hash: "foobar", Expires: 0}})

		// execute
		err := sut.CleanUp(ctx)

		// assert
		require.NoError(t, err)
	})

	t.Run("fails when store returns error", func(t *testing.T) {
		t.Parallel()

		// data

		// setup
		sessionStoreSpy := util.NewSpy()
		sessionStoreSpy.Register("ReadForWrite", 0, assert.AnError)

		sut := setup(t, sessionStoreSpy, nil)

		// execute
		err := sut.CleanUp(ctx)
		require.Error(t, err)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestSession_Get(t *testing.T) {
	t.Parallel()

	unusedSpy := util.NewSpy()
	ctx := context.Background()

	setup := func(t *testing.T, sessionStoreSpy *util.Spy, sessionData repo.SessionModelMap) *service.Session {
		t.Helper()

		sessionStore := store.NewInMemory(sessionStoreSpy)
		err := sessionStore.Marshal(ctx, sessionData)
		require.NoError(t, err)

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

		factory.SetStore(sessionStore, compose.SessionStore)

		return factory.CreateSessionService()
	}

	t.Run("fails when the session is expired", func(t *testing.T) {
		t.Parallel()

		// data
		userName := gofakeit.Name()
		sessionHash := gofakeit.UUID()
		sessionData := repo.SessionModelMap{
			userName: {
				Hash:    sessionHash,
				Expires: time.Now().Add(-time.Hour).Unix(),
			},
		}

		// setup
		sut := setup(t, unusedSpy, sessionData)

		// execute
		newSession, err := sut.Get(ctx, userName, sessionHash)
		require.Error(t, err)
		require.Empty(t, newSession)

		// assert
		assert.ErrorIs(t, err, apperr.ErrNotFound)
	})
}
