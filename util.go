package routeit

func stripDuplicates[T comparable](in []T) []T {
	indices := map[T]int{}
	i := 0
	for _, v := range in {
		if _, found := indices[v]; found {
			continue
		}
		indices[v] = i
		i++
	}

	out := make([]T, len(indices))
	for val, i := range indices {
		out[i] = val
	}

	return out
}
