package password_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peteraba/cloudy-files/password"
)

func FuzzBcrypt(f *testing.F) {
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
			t.Skip()
		}

		hash, err := sut.Hash(orig)
		require.NoError(t, err)

		err = sut.Check(orig, hash)
		require.NoError(t, err)
	})
}

func TestBcrypt_Check(t *testing.T) {
	t.Parallel()

	t.Run("password is incorrect", func(t *testing.T) {
		t.Parallel()

		stubPassword := "password"
		stubHash := "password"

		sut := password.NewBcrypt()

		err := sut.Check(stubPassword, stubHash)
		require.Error(t, err)

		assert.ErrorContains(t, err, "password is incorrect")
	})
}
