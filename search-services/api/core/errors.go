package core

import "errors"

var ErrBadArguments = errors.New("arguments are not acceptable")
var ErrAlreadyExists = errors.New("resource or task already exists")
var ErrTooLarge = errors.New("too large")
var ErrInternalServerError = errors.New("internal server error")
var ErrNotFound = errors.New("resource is not found")
