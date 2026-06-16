package slicex

func Map[S ~[]E, E, T any](s S, mapFn func(E) T) []T {
	if s == nil {
		return nil
	}

	o := make([]T, len(s))
	for i, v := range s {
		o[i] = mapFn(v)
	}

	return o
}
