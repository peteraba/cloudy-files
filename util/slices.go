package util

// This was an educated estimation based on benchmarks.
// Tested only on a 2024 MacBook Pro.
const countWhereMapBuildingIsFaster = 8_000

// Intersection returns the intersection of two slices.
func Intersection(sliceOne, sliceTwo []string) []string {
	if len(sliceOne)*len(sliceTwo) < countWhereMapBuildingIsFaster {
		return IntersectionSmall(sliceOne, sliceTwo)
	}

	return IntersectionLarge(sliceOne, sliceTwo)
}

// IntersectionSmall returns the intersection of two slices.
// This function is faster than IntersectionLarge for small slices.
func IntersectionSmall(sliceOne, sliceTwo []string) []string {
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

// IntersectionLarge returns the intersection of two slices.
// This function is faster than IntersectionSmall for large slices.
// This is *obviously* an overkill here, but it's a fun exercise.
func IntersectionLarge(sliceOne, sliceTwo []string) []string {
	lookUpTable := make(map[string]struct{})
	for _, v := range sliceOne {
		lookUpTable[v] = struct{}{}
	}

	var result []string

	for _, v := range sliceTwo {
		if _, ok := lookUpTable[v]; ok {
			result = append(result, v)
		}
	}

	return result
}
