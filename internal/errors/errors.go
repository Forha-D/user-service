package errors

import "errors"

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidInput      = errors.New("invalid input")
	ErrInternalServer    = errors.New("internal server error")
	ErrMissingEnv        = errors.New("missing required environment variable")
	ErrInvalidConfig     = errors.New("invalid configuration")
)
