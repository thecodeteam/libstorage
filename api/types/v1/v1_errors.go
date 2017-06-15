package v1

import "errors"

// ErrInvalidContext indicates an invalid object was provided as
// the argument to a Context parameter.
var ErrInvalidContext = errors.New("invalid context type")

// ErrInvalidConfig indicates an invalid object was provided as
// the argument to a Config parameter.
var ErrInvalidConfig = errors.New("invalid config type")

// ErrInvalidOp indicates an invalid operation ID.
var ErrInvalidOp = errors.New("invalid op")

// ErrInvalidArgsLen indicates the operation received an incorrect
// number of arguments.
var ErrInvalidArgsLen = errors.New("invalid args len")

// ErrInvalidArgs indicates the operation received one or more
// invalid arguments.
var ErrInvalidArgs = errors.New("invalid args")
