package service_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peteraba/cloudy-files/compose"
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

		factory := compose.NewFactory()

		factory.SetSessionStore(store.NewInMemoryFile(sessionStoreSpy))
		factory.SetUserStore(store.NewInMemoryFile(userStoreSpy))

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
		result, err := sut.Check(ctx, stubPassword, stubHash)
		require.Error(t, err)

		// assert
		assert.False(t, result)
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("session is invalid without a login", func(t *testing.T) {
		t.Parallel()

		// data
		stubName := gofakeit.Name()
		stubEmail := gofakeit.Email()
		stubPassword := gofakeit.Password(true, true, true, true, false, 16)
		stubAccess := []string{gofakeit.Adverb(), gofakeit.Adverb()}
		stubHash := gofakeit.UUID()

		// setup
		sut, userService := setup(t, unusedSpy, unusedSpy)

		err := userService.Create(ctx, stubName, stubEmail, stubPassword, false, stubAccess)
		require.NoError(t, err)

		// execute
		exists, err := sut.Check(ctx, stubName, stubHash)

		// assert
		require.NoError(t, err)
		assert.False(t, exists)
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

		err := userService.Create(ctx, stubName, stubEmail, stubPassword, false, stubAccess)
		require.NoError(t, err)

		sessionHash, err := userService.Login(ctx, stubName, stubPassword)
		require.NoError(t, err)
		require.NotEmpty(t, sessionHash)

		// execute
		exists, err := sut.Check(ctx, stubName, sessionHash)

		// assert
		require.NoError(t, err)
		assert.True(t, exists)
	})
}

func TestSession_CleanUp(t *testing.T) {
	t.Parallel()

	unusedSpy := util.NewSpy()
	ctx := context.Background()

	setup := func(t *testing.T, sessionStoreSpy *util.Spy, sessionData []byte) *service.Session {
		t.Helper()

		sessionStore := store.NewInMemoryFile(sessionStoreSpy)
		err := sessionStore.Write(ctx, sessionData)
		require.NoError(t, err)

		factory := compose.NewFactory()

		factory.SetSessionStore(sessionStore)

		return factory.CreateSessionService()
	}

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

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// setup
		storeData := []byte(`{"peter":{"hash":"foobar","expires":0}}`)

		sut := setup(t, unusedSpy, storeData)

		// execute
		err := sut.CleanUp(ctx)

		// assert
		require.NoError(t, err)
	})
}
