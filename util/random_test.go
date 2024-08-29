package util_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peteraba/cloudy-files/apperr"
	"github.com/peteraba/cloudy-files/util"
)

func TestRandomHex(t *testing.T) {
	t.Parallel()

	t.Run("should return an error if the random generation fails", func(t *testing.T) {
		t.Parallel()

		// execute
		_, err := util.RandomHex(-1)
		require.Error(t, err)

		// assert
		assert.ErrorIs(t, err, apperr.ErrInvalidArgument)
	})

	t.Run("should return a random hex string", func(t *testing.T) {
		t.Parallel()

		// execute
		hex, err := util.RandomHex(10)
		require.NoError(t, err)

		// assert
		assert.Len(t, hex, 10)
	})
}
