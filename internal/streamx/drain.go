package streamx

// Drain consumes a stream completely, returning all events as a slice.
// Iteration stops on the first error; collected events up to that point
// are returned along with the error.
func Drain[T any](stream Stream[T]) ([]T, error) {
	var result []T
	for v, err := range stream {
		if err != nil {
			return result, err
		}
		result = append(result, v)
	}
	return result, nil
}
