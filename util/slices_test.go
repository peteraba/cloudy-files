package util_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"

	"github.com/peteraba/cloudy-files/util"
)

func BenchmarkIntersectionSmall(b *testing.B) {
	slice150 := make([]string, 150)
	gofakeit.Slice(&slice150)

	b.Run("intersection30", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.IntersectionSmall(slice150[:30], slice150[10:40])
		}
	})

	b.Run("intersection70", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.IntersectionSmall(slice150[:75], slice150[50:120])
		}
	})

	b.Run("intersection125", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.IntersectionSmall(slice150[25:], slice150[:125])
		}
	})
}

func BenchmarkIntersectionLarge(b *testing.B) {
	slice150 := make([]string, 150)
	gofakeit.Slice(&slice150)

	b.Run("intersection30", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.IntersectionLarge(slice150[:30], slice150[10:40])
		}
	})

	b.Run("intersection70", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.IntersectionLarge(slice150[:75], slice150[50:120])
		}
	})

	b.Run("intersection125", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.IntersectionLarge(slice150[25:], slice150[:125])
		}
	})
}

func BenchmarkIntersection(b *testing.B) {
	slice150 := make([]string, 150)
	gofakeit.Slice(&slice150)

	b.Run("intersection30", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.Intersection(slice150[:30], slice150[10:40])
		}
	})

	b.Run("intersection70", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.Intersection(slice150[:75], slice150[50:120])
		}
	})

	b.Run("intersection125", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.Intersection(slice150[25:], slice150[:125])
		}
	})
}

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

func TestIntersectionSmall(t *testing.T) {
	t.Parallel()

	var (
		sliceOne = []string{"a", "b", "c", "d", "e"}
		sliceTwo = []string{"c", "d", "e", "f", "g"}
	)

	intersection := util.IntersectionSmall(sliceOne, sliceTwo)

	assert.Equal(t, []string{"c", "d", "e"}, intersection)
}

func TestIntersectionLarge(t *testing.T) {
	t.Parallel()

	var (
		sliceOne = []string{"a", "b", "c", "d", "e"}
		sliceTwo = []string{"c", "d", "e", "f", "g"}
	)

	intersection := util.IntersectionLarge(sliceOne, sliceTwo)

	assert.Equal(t, []string{"c", "d", "e"}, intersection)
}
