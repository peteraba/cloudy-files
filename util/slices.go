package util

func Intersection(a, b []string) []string {
	if len(a) < 20 && len(b) < 20 {
		return intersectionSmall(a, b)
	}

	return intersectionLarge(a, b)
}

func intersectionSmall(a, b []string) []string {
	intersection := make([]string, 0)

	for _, v := range a {
		for _, w := range b {
			if v == w {
				intersection = append(intersection, v)
			}
		}
	}

	return intersection
}

func intersectionLarge(sliceOne, sliceTwo []string) []string {
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
