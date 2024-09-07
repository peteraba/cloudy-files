package service_test

import (
	"context"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peteraba/cloudy-files/appconfig"
	"github.com/peteraba/cloudy-files/apperr"
	"github.com/peteraba/cloudy-files/compose"
	composeTest "github.com/peteraba/cloudy-files/compose/test"
	"github.com/peteraba/cloudy-files/password"
	"github.com/peteraba/cloudy-files/repo"
	"github.com/peteraba/cloudy-files/service"
	"github.com/peteraba/cloudy-files/store"
	"github.com/peteraba/cloudy-files/util"
)

func TestUser_Create_and_Login(t *testing.T) {
	t.Parallel()

	unusedSpy := util.NewSpy() // DO NOT USE !!!
	ctx := context.Background()

	setup := func(t *testing.T, userStoreSpy, sessionStoreSpy *util.Spy, userData repo.UserModelMap, sessionData repo.SessionModelMap) *service.User {
		t.Helper()

		userStore := store.NewInMemory(userStoreSpy)
		err := userStore.Marshal(ctx, userData)
		require.NoError(t, err)

		sessionStore := store.NewInMemory(sessionStoreSpy)
		err = sessionStore.Marshal(ctx, sessionData)
		require.NoError(t, err)

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

		factory.SetStore(userStore, compose.UserStore)
		factory.SetStore(sessionStore, compose.SessionStore)

		return factory.CreateUserService()
	}

	t.Run("fail to create user when passing is not OK", func(t *testing.T) {
		t.Parallel()

		// data
		stubName := gofakeit.Name()
		stubEmail := gofakeit.Email()
		stubPassword := strings.Repeat("foobar", 20)
		stubAccess := []string{gofakeit.Adverb(), gofakeit.Adverb()}

		// setup
		sut := setup(t, unusedSpy, unusedSpy, nil, nil)

		// execute
		userModel, err := sut.Create(ctx, stubName, stubEmail, stubPassword, false, stubAccess)
		require.Error(t, err)
		require.Empty(t, userModel)

		// assert
		assert.ErrorContains(t, err, "password is not OK")
	})

	t.Run("fail to create user when user store fails to read", func(t *testing.T) {
		t.Parallel()

		// data
		stubName := gofakeit.Name()
		stubEmail := gofakeit.Email()
		stubPassword := gofakeit.Password(true, true, true, true, false, 16)
		stubAccess := []string{gofakeit.Adverb(), gofakeit.Adverb()}

		// setup
		userStoreSpy := (util.NewSpy()).Register("ReadForWrite", 0, assert.AnError)

		sut := setup(t, userStoreSpy, unusedSpy, nil, nil)

		// execute
		userModel, err := sut.Create(ctx, stubName, stubEmail, stubPassword, false, stubAccess)
		require.Error(t, err)
		require.Empty(t, userModel)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("fail logging in with wrong password", func(t *testing.T) {
		t.Parallel()

		// data
		stubName := gofakeit.Name()
		stubEmail := gofakeit.Email()
		stubPassword := gofakeit.Password(true, true, true, true, false, 16)
		stubAccess := []string{gofakeit.Adverb(), gofakeit.Adverb()}

		// setup
		sut := setup(t, unusedSpy, unusedSpy, repo.UserModelMap{}, repo.SessionModelMap{})

		userModel, err := sut.Create(ctx, stubName, stubEmail, stubPassword, false, stubAccess)
		require.NoError(t, err)
		require.NotEmpty(t, userModel)

		// extra
		wrongPassword := stubPassword + " "
		err = sut.CheckPassword(ctx, stubName, wrongPassword)
		require.Error(t, err)

		// execute
		sessionHash, err := sut.Login(ctx, stubName, wrongPassword)
		require.Error(t, err)
		require.Empty(t, sessionHash)

		// assert
		assert.ErrorContains(t, err, "password does not match")
	})

	t.Run("login fails if user can not be found", func(t *testing.T) {
		t.Parallel()

		// data
		stubName := gofakeit.Name()
		stubPassword := ""

		// setup
		sut := setup(t, unusedSpy, unusedSpy, nil, nil)

		// execute
		hash, err := sut.Login(ctx, stubName, stubPassword)
		require.Error(t, err)
		require.Empty(t, hash)

		// assert
		assert.ErrorIs(t, err, apperr.ErrNotFound)
	})

	t.Run("non-admin user can log in", func(t *testing.T) {
		t.Parallel()

		// data
		stubName := gofakeit.Name()
		stubEmail := gofakeit.Email()
		stubPassword := gofakeit.Password(true, true, true, true, false, 16)
		stubAccess := []string{gofakeit.Adverb(), gofakeit.Adverb()}

		// setup
		sut := setup(t, unusedSpy, unusedSpy, repo.UserModelMap{}, repo.SessionModelMap{})

		userModel, err := sut.Create(ctx, stubName, stubEmail, stubPassword, false, stubAccess)
		require.NoError(t, err)
		require.NotEmpty(t, userModel)

		// extra
		err = sut.CheckPassword(ctx, stubName, stubPassword)
		require.NoError(t, err)

		// execute
		sessionHash, err := sut.Login(ctx, stubName, stubPassword)
		require.NoError(t, err)

		// assert
		assert.NotEmpty(t, sessionHash)
	})

	t.Run("can create an admin user and user can log in", func(t *testing.T) {
		t.Parallel()

		// data
		stubName := gofakeit.Name()
		stubEmail := gofakeit.Email()
		stubPassword := gofakeit.Password(true, true, true, true, false, 16)

		// setup
		sut := setup(t, unusedSpy, unusedSpy, repo.UserModelMap{}, repo.SessionModelMap{})

		userModel, err := sut.Create(ctx, stubName, stubEmail, stubPassword, true, []string{})
		require.NoError(t, err)
		require.NotEmpty(t, userModel)

		// extra
		err = sut.CheckPassword(ctx, stubName, stubPassword)
		require.NoError(t, err)

		// execute
		sessionHash, err := sut.Login(ctx, stubName, stubPassword)
		require.NoError(t, err)

		// assert
		assert.NotEmpty(t, sessionHash)
	})

	t.Run("password can be checked", func(t *testing.T) {
		t.Parallel()

		// data
		stubPassword := gofakeit.Password(true, true, true, true, false, 16)

		// setup
		sut := setup(t, unusedSpy, unusedSpy, repo.UserModelMap{}, repo.SessionModelMap{})
		passwordHash, err := sut.HashPassword(ctx, stubPassword)
		require.NoError(t, err)

		// execute
		err = sut.CheckPasswordHash(ctx, stubPassword, passwordHash)
		require.NoError(t, err)
	})
}

func TestUser_CheckPassword(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	setup := func(t *testing.T, userStoreSpy *util.Spy, userData repo.UserModelMap) *service.User {
		t.Helper()

		userStore := store.NewInMemory(userStoreSpy)
		err := userStore.Marshal(ctx, userData)
		require.NoError(t, err)

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

		factory.SetStore(userStore, compose.UserStore)

		return factory.CreateUserService()
	}

	t.Run("fail if user store fails to read", func(t *testing.T) {
		t.Parallel()

		// data
		stubName := gofakeit.Name()
		stubPassword := gofakeit.Password(true, true, true, true, false, 16)

		// setup
		userStoreSpy := (util.NewSpy()).Register("Read", 0, assert.AnError)

		sut := setup(t, userStoreSpy, repo.UserModelMap{})

		// execute
		err := sut.CheckPassword(ctx, stubName, stubPassword)
		require.Error(t, err)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestUser_HashPassword(t *testing.T) {
	t.Parallel()

	unusedSpy := util.NewSpy() // DO NOT USE !!!
	ctx := context.Background()

	setup := func(t *testing.T, hasherSpy *util.Spy) *service.User {
		t.Helper()

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

		hasher := password.NewDummyHasher(hasherSpy)
		factory.SetHasher(hasher)

		return factory.CreateUserService()
	}

	t.Run("fail if password is not OK", func(t *testing.T) {
		t.Parallel()

		// data
		stubPassword := strings.Repeat("foobar", 20)

		// setup
		sut := setup(t, unusedSpy)

		// execute
		hash, err := sut.HashPassword(ctx, stubPassword)
		require.Error(t, err)

		// assert
		assert.ErrorContains(t, err, "password is not OK")
		assert.Empty(t, hash)
	})

	t.Run("fail if password is not OK", func(t *testing.T) {
		t.Parallel()

		// data
		stubPassword := "foobarFOOBAR123"

		// setup
		hasherSpy := util.NewSpy()
		hasherSpy.Register("Hash", 0, assert.AnError, stubPassword)
		sut := setup(t, hasherSpy)

		// execute
		hash, err := sut.HashPassword(ctx, stubPassword)
		require.Error(t, err)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
		assert.Empty(t, hash)
	})
}

func TestUser_CheckPasswordHash(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	setup := func(t *testing.T, hasherSpy *util.Spy) *service.User {
		t.Helper()

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

		hasher := password.NewDummyHasher(hasherSpy)
		factory.SetHasher(hasher)

		return factory.CreateUserService()
	}

	t.Run("fail if hasher fails", func(t *testing.T) {
		t.Parallel()

		// data
		stubPassword := "foobarFOOBAR123"
		stubHash := "my-hash-foo-bar-123"

		// setup
		hasherSpy := util.NewSpy()
		hasherSpy.Register("Check", 0, assert.AnError, stubPassword, stubHash)
		sut := setup(t, hasherSpy)

		// execute
		err := sut.CheckPasswordHash(ctx, stubPassword, stubHash)
		require.Error(t, err)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})
}
