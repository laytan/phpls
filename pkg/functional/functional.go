// Package functional provides functional utilities for commonly done operations.
package functional

// MapFilter applies the mapper function to each entry in the slice and returns a new
// slice with the results.
//
// If mapper returns the default value for the R type, the result is not added
// to the returned slice.
func MapFilter[V any, R comparable](slice []V, mapper func(entry V) R) []R {
	res := make([]R, 0, len(slice))
	var defaultVal R

	for _, v := range slice {
		if mapped := mapper(v); mapped != defaultVal {
			res = append(res, mapped)
		}
	}

	return res
}

// Map applies the mapper function to each entry in the slice and returns a new
// slice with the results.
func Map[V any, R any](slice []V, mapper func(entry V) R) []R {
	res := make([]R, 0, len(slice))
	for _, v := range slice {
		res = append(res, mapper(v))
	}

	return res
}

type stringer interface {
	String() string
}

// ToString is a function useful to pass into a mapper, to map all values to a string.
func ToString[T stringer](v T) string {
	return v.String()
}
