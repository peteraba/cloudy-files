package util_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"

	"github.com/peteraba/cloudy-files/util"
)

var slice300 = make([]string, 300)

func init() {
	gofakeit.Slice(&slice300)
}

func BenchmarkIntersectionSmall(b *testing.B) {
	b.Run("intersection30", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.IntersectionSmall(slice300[:30], slice300[:30])
		}
	})

	b.Run("intersection75", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.IntersectionSmall(slice300[:75], slice300[:75])
		}
	})

	b.Run("intersection100", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.IntersectionSmall(slice300[:100], slice300[:100])
		}
	})

	b.Run("intersection100+", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.IntersectionSmall(slice300[:75], slice300[:150])
		}
	})

	b.Run("intersection150", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.IntersectionSmall(slice300[:150], slice300[:150])
		}
	})
}

func BenchmarkIntersectionLarge(b *testing.B) {
	b.Run("intersection30", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.IntersectionLarge(slice300[:30], slice300[:30])
		}
	})

	b.Run("intersection75", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.IntersectionLarge(slice300[:75], slice300[:75])
		}
	})

	b.Run("intersection100", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.IntersectionLarge(slice300[:100], slice300[:100])
		}
	})

	b.Run("intersection100+", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.IntersectionLarge(slice300[:75], slice300[:150])
		}
	})

	b.Run("intersection150", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.IntersectionLarge(slice300[:150], slice300[:150])
		}
	})
}

func BenchmarkIntersection(b *testing.B) {
	b.Run("intersection30", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.Intersection(slice300[:30], slice300[:30])
		}
	})

	b.Run("intersection75", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.Intersection(slice300[:75], slice300[:75])
		}
	})

	b.Run("intersection100", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.Intersection(slice300[:100], slice300[:100])
		}
	})

	b.Run("intersection100+", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.Intersection(slice300[:75], slice300[:150])
		}
	})

	b.Run("intersection150", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.Intersection(slice300[:150], slice300[:150])
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

		intersection := util.Intersection(slice300[:20], slice300[10:30])
		reverse := util.Intersection(slice300[10:30], slice300[:20])

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
