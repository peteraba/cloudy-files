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
			util.IntersectionSmall(slice150[:30], slice150[:30])
		}
	})

	b.Run("intersection75", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.IntersectionSmall(slice150[:75], slice150[:75])
		}
	})

	b.Run("intersection100", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.IntersectionSmall(slice150[:100], slice150[:100])
		}
	})

	b.Run("intersection100+", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.IntersectionSmall(slice150[:75], slice150[:150])
		}
	})

	b.Run("intersection150", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.IntersectionSmall(slice150, slice150)
		}
	})
}

func BenchmarkIntersectionLarge(b *testing.B) {
	slice150 := make([]string, 150)
	gofakeit.Slice(&slice150)

	b.Run("intersection30", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.IntersectionLarge(slice150[:30], slice150[:30])
		}
	})

	b.Run("intersection75", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.IntersectionLarge(slice150[:75], slice150[:75])
		}
	})

	b.Run("intersection100", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.IntersectionLarge(slice150[:100], slice150[:100])
		}
	})

	b.Run("intersection100+", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.IntersectionLarge(slice150[:75], slice150[:150])
		}
	})

	b.Run("intersection150", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.IntersectionLarge(slice150, slice150)
		}
	})
}

func BenchmarkIntersection(b *testing.B) {
	slice150 := make([]string, 150)
	gofakeit.Slice(&slice150)

	b.Run("intersection30", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.Intersection(slice150[:30], slice150[:30])
		}
	})

	b.Run("intersection75", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.Intersection(slice150[:75], slice150[:75])
		}
	})

	b.Run("intersection100", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.Intersection(slice150[:100], slice150[:100])
		}
	})

	b.Run("intersection100+", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.Intersection(slice150[:75], slice150[:150])
		}
	})

	b.Run("intersection150", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.Intersection(slice150, slice150)
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

		slice30 := make([]string, 30)
		gofakeit.Slice(&slice30)

		intersection := util.Intersection(slice30[:20], slice30[10:])
		reverse := util.Intersection(slice30[10:], slice30[:20])

		assert.Equal(t, intersection, reverse)
		assert.GreaterOrEqual(t, len(intersection), 10)
		assert.LessOrEqual(t, len(intersection), 13) // This is somewhat made up, but even 3 extra match is highly unlikely
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
