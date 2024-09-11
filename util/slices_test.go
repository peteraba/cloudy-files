package util_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"

	"github.com/peteraba/cloudy-files/util"
)

func TestHasIntersection(t *testing.T) {
	t.Parallel()

	t.Run("simple", func(t *testing.T) {
		t.Parallel()

		var (
			sliceOne = []string{"a", "b", "c", "d", "e"}
			sliceTwo = []string{"c", "d", "e", "f", "g"}
		)

		assert.True(t, util.HasIntersection(sliceOne, sliceTwo))
	})

	t.Run("large", func(t *testing.T) {
		t.Parallel()

		slice200 := make([]string, 200)
		gofakeit.Slice(&slice200)

		assert.True(t, util.HasIntersection(slice200[:125], slice200[75:]))
	})
}
