package utils

// Ptr returns a pointer to the passed value.
func Ptr[T any](v T) *T {
	return &v
}

// Deref returns the value pointed to by p if non-nil, otherwise fallback.
func Deref[T any](p *T, fallback T) T {
	if p != nil {
		return *p
	}
	return fallback
}

// DerefOrZero returns the value pointed to by p if non-nil, otherwise the zero value of T.
func DerefOrZero[T any](p *T) T {
	var zero T
	if p != nil {
		return *p
	}
	return zero
}
