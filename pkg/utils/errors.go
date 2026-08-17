package utils

import (
	"errors"
	"fmt"
)

type ClientError struct{ Err error }

func (e ClientError) Error() string { return e.Err.Error() }
func (e ClientError) Unwrap() error { return e.Err }

type AuthorizationError struct{ Err error }

func (e AuthorizationError) Error() string { return e.Err.Error() }
func (e AuthorizationError) Unwrap() error { return e.Err }

func NewClientErrorf(format string, args ...any) error {
	return ClientError{Err: fmt.Errorf(format, args...)}
}

func NewAuthorizationError(err error) error {

	if err == nil {
		return nil
	}

	return AuthorizationError{Err: err}
}

func IsClientError(err error) bool {
	var target ClientError
	return errors.As(err, &target)
}

func IsAuthorizationError(err error) bool {
	var target AuthorizationError
	return errors.As(err, &target)
}
