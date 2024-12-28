package main

// helper types

// Future represents a placeholder for a result that will be available in the future.
type Future[T any] struct {
	// Implementation details of Future.
	Result T
	Err    error
}
