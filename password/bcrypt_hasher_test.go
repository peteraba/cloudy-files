package password_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peteraba/cloudy-files/password"
)

func FuzzBcrypt(f *testing.F) {
	ctx := context.Background()

	testCases := []string{
		"password",
		"97TZPRZFGZFX9g",
		"Supreme executive power derives from a mandate from the masses",
	}

	for _, tc := range testCases {
		f.Add(tc) // Use f.Add to provide a seed corpus
	}

	sut := password.NewBcryptWithCost(0)

	f.Fuzz(func(t *testing.T, orig string) {
		if len(orig) > 72 {
			orig = string([]byte(orig)[:72])
		}

		hash, err := sut.Hash(ctx, orig)
		require.NoError(t, err)

		err = sut.Check(ctx, orig, hash)
		require.NoError(t, err)
	})
}

func TestBcrypt_Hash(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("fail to hash too long password", func(t *testing.T) {
		t.Parallel()

		// setup
		stubPassword := strings.Repeat("foobar", 20)

		sut := password.NewBcryptHasher()

		// execute
		_, err := sut.Hash(ctx, stubPassword)
		require.Error(t, err)

		// assert
		assert.ErrorContains(t, err, "password length exceeds")
	})

	t.Run("can hash crazy unicode characters", func(t *testing.T) {
		t.Parallel()

		// setup
		stubPassword := "ő✈♸⛄" //nolint:gosec // This is an example password, no need to worry.

		sut := password.NewBcryptHasher()

		// execute
		hash, err := sut.Hash(ctx, stubPassword)
		require.NoError(t, err)

		// assert
		assert.NotEmpty(t, hash)
		assert.NoError(t, sut.Check(ctx, stubPassword, hash))
	})
}

func TestBcrypt_Check(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("password is incorrect", func(t *testing.T) {
		t.Parallel()

		// data
		stubPassword := "password"
		stubHash := "password"

		// setup
		sut := password.NewBcryptHasher()

		// execute
		err := sut.Check(ctx, stubPassword, stubHash)
		require.Error(t, err)

		// assert
		assert.ErrorContains(t, err, "password is incorrect")
	})
}
