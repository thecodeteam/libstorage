package v1

// Driver the interface for types that are drivers.
type Driver interface {

	// Do does the operation specified by the opID.
	Do(
		ctx interface{},
		opID uint64,
		args ...interface{}) (result interface{}, err error)

	// Init initializes the driver instance.
	Init(ctx, config interface{}) error

	// Name returns the name of the driver.
	Name() string

	// Supports returns a mask of the operations supported by the driver.
	Supports() uint64
}

const (
	// DtStorage indicates the storage driver type.
	DtStorage uint8 = 1 << iota

	// DtIntegration indicates the integration driver type.
	DtIntegration
)
