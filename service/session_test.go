package service_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peteraba/cloudy-files/compose"
	"github.com/peteraba/cloudy-files/store"
)

func TestSession_Check(t *testing.T) {
	t.Parallel()

	factory := compose.NewFactory()

	factory.SetSessionStore(store.NewInMemoryFile())
	factory.SetUserStore(store.NewInMemoryFile())
	factory.SetSessionStore(store.NewInMemoryFile())

	userService := factory.CreateUserService()

	sut := factory.CreateSessionService()

	t.Run("session is invalid without a login", func(t *testing.T) {
		t.Parallel()

		// setup
		stubName := gofakeit.Name()
		stubEmail := gofakeit.Email()
		stubPassword := gofakeit.Password(true, true, true, true, false, 16)
		stubAccess := []string{gofakeit.Adverb(), gofakeit.Adverb()}
		stubHash := gofakeit.UUID()

		err := userService.Create(stubName, stubEmail, stubPassword, false, stubAccess)
		require.NoError(t, err)

		// execute
		exists, err := sut.Check(stubName, stubHash)
		require.NoError(t, err)

		// assert
		assert.False(t, exists)
	})

	t.Run("logged in user has valid session", func(t *testing.T) {
		t.Parallel()

		// setup
		stubName := gofakeit.Name()
		stubEmail := gofakeit.Email()
		stubPassword := gofakeit.Password(true, true, true, true, false, 16)
		stubAccess := []string{gofakeit.Adverb(), gofakeit.Adverb()}

		err := userService.Create(stubName, stubEmail, stubPassword, false, stubAccess)
		require.NoError(t, err)

		sessionHash, err := userService.Login(stubName, stubPassword)
		require.NoError(t, err)
		require.NotEmpty(t, sessionHash)

		// execute
		exists, err := sut.Check(stubName, sessionHash)
		require.NoError(t, err)

		// assert
		assert.True(t, exists)
	})
}
