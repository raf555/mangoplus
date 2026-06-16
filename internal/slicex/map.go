package slicex

func Map[S ~[]E, O []T, E, T any](s S, mapFn func(E) T) O {
	if s == nil {
		return nil
	}

	o := make(O, len(s))
	for i, v := range s {
		o[i] = mapFn(v)
	}

	return o
}
