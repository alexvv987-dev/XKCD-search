package core

import "errors"

var ErrBadArguments = errors.New("arguments are not acceptable")
var ErrNotFound = errors.New("resource is not found")
var ErrTooLarge = errors.New("too large")
var ErrInternalServerError = errors.New("internal server error")
