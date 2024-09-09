package password_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peteraba/cloudy-files/apperr"
	"github.com/peteraba/cloudy-files/password"
	"github.com/peteraba/cloudy-files/util"
)

func TestDummyHasher_Hash(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	setup := func(t *testing.T) (*password.DummyHasher, *util.Spy) {
		t.Helper()

		spy := util.NewSpy()
		sut := password.NewDummyHasher(spy)

		return sut, spy
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// data
		passwordStub := "password"

		// setup
		sut, _ := setup(t)

		// execute
		hash, err := sut.Hash(ctx, passwordStub)
		require.NoError(t, err)

		err = sut.Check(ctx, passwordStub, hash)
		require.NoError(t, err)

		// assert
		assert.NotEqual(t, passwordStub, hash)
	})

	t.Run("fail if checked with incorrect hash", func(t *testing.T) {
		t.Parallel()

		// data
		passwordStub := "password"

		// setup
		sut, _ := setup(t)

		// execute
		err := sut.Check(ctx, passwordStub, passwordStub)
		require.Error(t, err)

		// assert
		assert.ErrorIs(t, err, apperr.ErrPasswordTooLong)
	})

	t.Run("fail if hashing fails", func(t *testing.T) {
		t.Parallel()

		// data
		passwordStub := "password"

		// setup
		sut, spy := setup(t)
		spy.Register("Hash", 0, assert.AnError, passwordStub)

		// execute
		hash, err := sut.Hash(ctx, passwordStub)
		require.Error(t, err)
		require.Empty(t, hash)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("fail if checking fails", func(t *testing.T) {
		t.Parallel()

		// data
		passwordStub := "password"
		hashStub := "hash"

		// setup
		sut, spy := setup(t)
		spy.Register("Check", 0, assert.AnError, passwordStub, hashStub)

		// execute
		err := sut.Check(ctx, passwordStub, hashStub)
		require.Error(t, err)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})
}
