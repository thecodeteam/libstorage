package v1

// InstanceID identifies a host to a remote storage platform.
type InstanceID interface {

	// GetID returns the ID of an instance.
	GetID() string

	// GetDriver returns the name of the driver that created the
	// instance ID.
	GetDriver() string

	// GetService returns the name of the storage service for which the
	// instance ID is valid..
	GetService() string

	// GetFields returns additional data about the object.
	GetFields() map[string]string
}

// Instance provides information about a storage object.
type Instance interface {

	// GetInstanceID returns an InstanceID, the ID of the instance to
	// which the object is connected.
	GetInstanceID() interface{}

	// GetName returns the name of the instance.
	GetName() string

	// GetProviderName returns the name of the provider that owns the
	// object.
	GetProviderName() string

	// GetRegion returns the region from which the object originates.
	GetRegion() string

	// GetFields are additional properties that can be defined for this
	// type.
	GetFields() map[string]string
}
