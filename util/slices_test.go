package util_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"

	"github.com/peteraba/cloudy-files/util"
)

func TestIntersection(t *testing.T) {
	t.Parallel()

	t.Run("simple", func(t *testing.T) {
		t.Parallel()

		var (
			sliceOne = []string{"a", "b", "c", "d", "e"}
			sliceTwo = []string{"c", "d", "e", "f", "g"}
		)

		intersection := util.Intersection(sliceOne, sliceTwo)

		assert.Equal(t, []string{"c", "d", "e"}, intersection)
	})

	t.Run("large", func(t *testing.T) {
		t.Parallel()

		slice200 := make([]string, 200)
		gofakeit.Slice(&slice200)

		intersection := util.Intersection(slice200[:125], slice200[75:])
		reverse := util.Intersection(slice200[75:], slice200[:125])

		assert.Equal(t, intersection, reverse)
		assert.GreaterOrEqual(t, len(intersection), 50)
		assert.LessOrEqual(t, len(intersection), 60) // This is somewhat made up, but even 10 extra match is highly unlikely
	})
}
