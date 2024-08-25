package util_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"

	"github.com/peteraba/cloudy-files/util"
)

func BenchmarkIntersectionSmall(b *testing.B) {
	var (
		slice30  = make([]string, 30)
		slice75  = make([]string, 75)
		slice100 = make([]string, 100)
		slice150 = make([]string, 150)
	)

	gofakeit.Slice(&slice30)
	gofakeit.Slice(&slice75)
	gofakeit.Slice(&slice100)
	gofakeit.Slice(&slice150)

	b.Run("intersection30", func(b *testing.B) {
		// b.ResetTimer()

		for i := 0; i < b.N; i++ {
			util.IntersectionSmall(slice30, slice30)
		}
	})

	b.Run("intersection75", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.IntersectionSmall(slice75, slice75)
		}
	})

	b.Run("intersection100", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.IntersectionSmall(slice100, slice100)
		}
	})

	b.Run("intersection100+", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.IntersectionSmall(slice75, slice150)
		}
	})

	b.Run("intersection150", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.IntersectionSmall(slice150, slice150)
		}
	})
}

func BenchmarkIntersectionLarge(b *testing.B) {
	var (
		slice30  = make([]string, 30)
		slice75  = make([]string, 75)
		slice100 = make([]string, 100)
		slice150 = make([]string, 150)
	)

	gofakeit.Slice(&slice30)
	gofakeit.Slice(&slice75)
	gofakeit.Slice(&slice100)
	gofakeit.Slice(&slice150)

	b.Run("intersection30", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.IntersectionLarge(slice30, slice30)
		}
	})

	b.Run("intersection75", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.IntersectionLarge(slice75, slice75)
		}
	})

	b.Run("intersection100", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.IntersectionLarge(slice100, slice100)
		}
	})

	b.Run("intersection100+", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.IntersectionLarge(slice75, slice150)
		}
	})

	b.Run("intersection150", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.IntersectionLarge(slice150, slice150)
		}
	})
}

func BenchmarkIntersection(b *testing.B) {
	var (
		slice30  = make([]string, 30)
		slice75  = make([]string, 75)
		slice100 = make([]string, 100)
		slice150 = make([]string, 150)
	)

	gofakeit.Slice(&slice30)
	gofakeit.Slice(&slice75)
	gofakeit.Slice(&slice100)
	gofakeit.Slice(&slice150)

	b.Run("intersection30", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.Intersection(slice30, slice30)
		}
	})

	b.Run("intersection75", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.Intersection(slice75, slice75)
		}
	})

	b.Run("intersection100", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.Intersection(slice100, slice100)
		}
	})

	b.Run("intersection100+", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			util.Intersection(slice75, slice150)
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

	var (
		sliceOne = []string{"a", "b", "c", "d", "e"}
		sliceTwo = []string{"c", "d", "e", "f", "g"}
	)

	intersection := util.Intersection(sliceOne, sliceTwo)

	assert.Equal(t, []string{"c", "d", "e"}, intersection)
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
