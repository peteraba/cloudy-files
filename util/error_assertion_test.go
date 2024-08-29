package util_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/peteraba/cloudy-files/util"
)

// Note: testing failures is hard due to the fact that helper would fail the test.
func TestErrorContains(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// setup
		assertFunc := util.ErrorIs(assert.AnError)

		// exercise
		result := assertFunc(t, assert.AnError)

		// assert
		assert.True(t, result)
	})
}

// Note: testing failures is hard due to the fact that helper would fail the test.
func TestErrorIs(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// setup
		assertFunc := util.ErrorContains(assert.AnError.Error())

		// exercise
		result := assertFunc(t, assert.AnError)

		// assert
		assert.True(t, result)
	})
}
