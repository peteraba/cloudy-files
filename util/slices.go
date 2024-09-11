package util

// HasIntersection checks if two slices have any common elements.
func HasIntersection(sliceOne, sliceTwo []string) bool {
	for _, v := range sliceOne {
		for _, w := range sliceTwo {
			if v == w {
				return true
			}
		}
	}

	return false
}
