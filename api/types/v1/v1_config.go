package v1

// Config is the interface that enables retrieving configuration information.
// The variations of the Get function, the Set, IsSet, and Scope functions
// all take an interface{} as their first parameter. However, the param must be
// either a string or a fmt.Stringer, otherwise the function will panic.
type Config interface {

	// GetString returns the value associated with the key as a string
	GetString(k interface{}) string

	// GetBool returns the value associated with the key as a bool
	GetBool(k interface{}) bool

	// GetStringSlice returns the value associated with the key as a string
	// slice.
	GetStringSlice(k interface{}) []string

	// GetInt returns the value associated with the key as an int
	GetInt(k interface{}) int

	// Get returns the value associated with the key
	Get(k interface{}) interface{}

	// Set sets an override value
	Set(k interface{}, v interface{})

	// IsSet returns a flag indicating whether or not a key is set.
	IsSet(k interface{}) bool
}
