package service_test

import (
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peteraba/cloudy-files/compose"
	"github.com/peteraba/cloudy-files/service"
	"github.com/peteraba/cloudy-files/store"
	"github.com/peteraba/cloudy-files/util"
)

func TestUser_Create_and_Login(t *testing.T) {
	t.Parallel()

	unusedSpy := util.NewSpy()

	setup := func(t *testing.T, userStoreSpy, sessionStoreSpy *util.Spy, userData, sessionData []byte) *service.User {
		t.Helper()

		userStore := store.NewInMemoryFile(userStoreSpy)
		err := userStore.Write(userData)
		require.NoError(t, err)

		sessionStore := store.NewInMemoryFile(sessionStoreSpy)
		err = sessionStore.Write(sessionData)
		require.NoError(t, err)

		factory := compose.NewFactory()

		factory.SetUserStore(userStore)
		factory.SetSessionStore(sessionStore)

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
		err := sut.Create(stubName, stubEmail, stubPassword, false, stubAccess)

		// assert
		require.Error(t, err)
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
		err := sut.Create(stubName, stubEmail, stubPassword, false, stubAccess)

		// assert
		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("fail to create user when user store reads invalid data", func(t *testing.T) {
		t.Parallel()

		// data
		stubName := gofakeit.Name()
		stubEmail := gofakeit.Email()
		stubPassword := gofakeit.Password(true, true, true, true, false, 16)
		stubAccess := []string{gofakeit.Adverb(), gofakeit.Adverb()}

		// setup
		sut := setup(t, unusedSpy, unusedSpy, []byte("invalid json"), nil)

		// execute
		err := sut.Create(stubName, stubEmail, stubPassword, false, stubAccess)

		// assert
		require.Error(t, err)
		assert.ErrorContains(t, err, "error unmarshaling data")
	})

	t.Run("logging in with wrong password does not work", func(t *testing.T) {
		t.Parallel()

		// data
		stubName := gofakeit.Name()
		stubEmail := gofakeit.Email()
		stubPassword := gofakeit.Password(true, true, true, true, false, 16)
		stubAccess := []string{gofakeit.Adverb(), gofakeit.Adverb()}

		// setup
		sut := setup(t, unusedSpy, unusedSpy, nil, nil)

		err := sut.Create(stubName, stubEmail, stubPassword, false, stubAccess)
		require.NoError(t, err)

		// extra
		wrongPassword := stubPassword + " "
		err = sut.CheckPassword(stubName, wrongPassword)
		require.Error(t, err)

		// execute
		sessionHash, err := sut.Login(stubName, wrongPassword)
		require.Error(t, err)
		require.Empty(t, sessionHash)
	})

	t.Run("non-admin user can log in", func(t *testing.T) {
		t.Parallel()

		// data
		stubName := gofakeit.Name()
		stubEmail := gofakeit.Email()
		stubPassword := gofakeit.Password(true, true, true, true, false, 16)
		stubAccess := []string{gofakeit.Adverb(), gofakeit.Adverb()}

		// setup
		sut := setup(t, unusedSpy, unusedSpy, nil, nil)

		err := sut.Create(stubName, stubEmail, stubPassword, false, stubAccess)
		require.NoError(t, err)

		// extra
		err = sut.CheckPassword(stubName, stubPassword)
		require.NoError(t, err)

		// execute
		sessionHash, err := sut.Login(stubName, stubPassword)
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
		sut := setup(t, unusedSpy, unusedSpy, nil, nil)

		err := sut.Create(stubName, stubEmail, stubPassword, true, []string{})
		require.NoError(t, err)

		// extra
		err = sut.CheckPassword(stubName, stubPassword)
		require.NoError(t, err)

		// execute
		sessionHash, err := sut.Login(stubName, stubPassword)
		require.NoError(t, err)

		// assert
		assert.NotEmpty(t, sessionHash)
	})

	t.Run("password can be checked", func(t *testing.T) {
		t.Parallel()

		// data
		stubPassword := gofakeit.Password(true, true, true, true, false, 16)

		// setup
		sut := setup(t, unusedSpy, unusedSpy, nil, nil)
		passwordHash, err := sut.HashPassword(stubPassword)
		require.NoError(t, err)

		// execute
		err = sut.CheckPasswordHash(stubPassword, passwordHash)
		require.NoError(t, err)
	})
}
