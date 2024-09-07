package util

// Intersection returns the intersection of two slices.
func Intersection(sliceOne, sliceTwo []string) []string {
	intersection := make([]string, 0)

	for _, v := range sliceOne {
		for _, w := range sliceTwo {
			if v == w {
				intersection = append(intersection, v)
			}
		}
	}

	return intersection
}
