package service_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peteraba/cloudy-files/compose"
	"github.com/peteraba/cloudy-files/store"
)

func TestUser_Create_and_Login(t *testing.T) {
	t.Parallel()

	factory := compose.NewFactory()

	factory.SetUserStore(store.NewInMemoryFile())
	factory.SetSessionStore(store.NewInMemoryFile())

	sut := factory.CreateUserService()

	t.Run("non-admin user can log in", func(t *testing.T) {
		t.Parallel()

		// setup
		stubName := gofakeit.Name()
		stubEmail := gofakeit.Email()
		stubPassword := gofakeit.Password(true, true, true, true, false, 16)
		stubAccess := []string{gofakeit.Adverb(), gofakeit.Adverb()}

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

		// setup
		stubName := gofakeit.Name()
		stubEmail := gofakeit.Email()
		stubPassword := gofakeit.Password(true, true, true, true, false, 16)

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

		// setup
		stubPassword := gofakeit.Password(true, true, true, true, false, 16)

		passwordHash, err := sut.HashPassword(stubPassword)
		require.NoError(t, err)

		// execute
		err = sut.CheckPasswordHash(stubPassword, passwordHash)
		require.NoError(t, err)
	})
}
