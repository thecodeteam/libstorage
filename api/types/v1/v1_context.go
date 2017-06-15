package v1

import "context"

// Context is an extension of the Go Context.
type Context interface {
	context.Context
}

// ContextLoggerFieldAware is used by types that will be logged by the
// Context logger. The key/value pair returned by the type is then emitted
// as part of  the Context's log entry.
type ContextLoggerFieldAware interface {

	// ContextLoggerField is the fields that is logged as part of a Context's
	// log entry.
	ContextLoggerField() (string, interface{})
}

// ContextLoggerFieldsAware is used by types that will be logged by the
// Context logger. The fields returned by the type are then emitted as part of
// the Context's log entry.
type ContextLoggerFieldsAware interface {

	// ContextLoggerFields are the fields that are logged as part of a
	// Context's log entry.
	ContextLoggerFields() map[string]interface{}
}
