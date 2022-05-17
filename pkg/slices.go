package pkg

func MapSlice[S, T any](s []S, mapFn func(S) T) []T {
	result := make([]T, len(s))
	for i := 0; i < len(s); i++ {
		result[i] = mapFn(s[i])
	}
	return result
}
