package v1

const (
	// SdNextDevice returns the next available device for the system
	// on which the function is executed for the storage platform
	// represented by the driver.
	//
	//     func(ctx) (string, error)
	SdNextDevice uint64 = 1 << iota

	// SdLocalDevices returns a map of the local devices for the system
	// on which the function is executed for the storage platform
	// represented by the driver.
	//
	//     func(ctx, LocalDeviceOpts) (LocalDevices, error)
	SdLocalDevices

	// SdLogin creates a new connection to the storage platform for the
	// provided context.
	//
	//     func(ctx) (interface{}, error)
	SdLogin

	// SdNextDeviceInfo returns information about the next device
	// workflow for the storage platform represented by the driver.
	//
	//     func(ctx) (NextDeviceInfo, error)
	SdNextDeviceInfo

	// SdType returns the type of storage provided by the storage
	// platform represented by the driver.
	//
	//     func(ctx) (uint8, error)
	SdType

	// SdInstanceID returns the ID of the system on which the function
	// is executed for the storage platform represented by the driver.
	//
	//     func(ctx) (InstanceID, error)
	SdInstanceID

	// SdInstanceIDInspect returns the complete Instance information for
	// an InstanceID for the storage platform represented by the driver.
	//
	//     func(ctx, InstanceID) (Instance, error)
	SdInstanceIDInspect

	// SdVolumes returns a list of volumes for the storage platform represented
	// by the storage driver.
	//
	//     func(ctx, VolumesOpts) ([]Volume, error)
	SdVolumes

	// SdVolumeInspect inspects a single volume.
	//
	//     func(
	//         ctx,
	//         volumeID string,
	//         VolumeInspectOpts) (Volume, error)
	SdVolumeInspect

	// SdVolumeInspectByName inspects a single volume by name.
	//
	//     func(
	//         ctx,
	//         volumeName string,
	//         VolumeInspectOpts) (Volume, error)
	SdVolumeInspectByName

	// SdVolumeCreate creates a new volume.
	//
	//     func(
	//         ctx,
	//         name string,
	//         VolumeCreateOpts) (Volume, error)
	SdVolumeCreate

	// SdVolumeCreateFromSnapshot creates a new volume from an existing snapshot.
	//
	//     func(
	//         ctx,
	//         snapshotID, volumeName string,
	//         VolumeCreateOpts) (Volume, error)
	SdVolumeCreateFromSnapshot

	// SdVolumeCopy copies an existing volume.
	//
	//     func(
	//         ctx,
	//         volumeID, volumeName string,
	//         Store) (Volume, error)
	SdVolumeCopy

	// SdVolumeSnapshot snapshots a volume.
	//
	//     func(
	//         ctx,
	//         volumeID, snapshotName string,
	//         Store) (Snapshot, error)
	SdVolumeSnapshot

	// SdVolumeRemove removes a volume.
	//
	//     func(
	//         ctx,
	//         volumeID string,
	//         VolumeRemoveOpts) (nil, error)
	SdVolumeRemove

	// SdVolumeAttach attaches a volume and provides a token clients can use
	// to validate that device has appeared locally.
	//
	//     func(
	//         ctx,
	//         volumeID string,
	//         VolumeAttachOpts) (VolumeAttachResult, error)
	SdVolumeAttach

	// SdVolumeDetach detaches a volume.
	//
	//     func(
	//         ctx,
	//         volumeID string,
	//         VolumeDetachOpts) (Volume, error)
	SdVolumeDetach

	// SdSnapshots returns all volumes or a filtered list of snapshots.
	//
	//     func(ctx, Store) ([]Snapshot, error)
	SdSnapshots

	// SdSnapshotInspect inspects a single snapshot.
	//
	//     func(
	//         ctx,
	//         snapshotID string,
	//         Store) (Snapshot, error)
	SdSnapshotInspect

	// SdSnapshotCopy copies an existing snapshot.
	//
	//     func(
	//         ctx,
	//         snapshotID, snapshotName, destinationID string,
	//         Store) (Snapshot, error)
	SdSnapshotCopy

	// SdSnapshotRemove removes a snapshot.
	//
	//     func(
	//         ctx,
	//         snapshotID string,
	//         Store) (nil, error)
	SdSnapshotRemove
)

const (

	// SdInstOpMin is the first storage driver instance operation.
	SdInstOpMin = SdInstanceID

	// SdInstOpMax is the last storage driver instance operation.
	SdInstOpMax = SdInstanceIDInspect

	// SdVolOpIDMin is the first storage driver volume operation.
	SdVolOpIDMin = SdVolumes

	// SdVolOpIDMax is the last storage driver volume operation.
	SdVolOpIDMax = SdVolumeDetach

	// SdSnapOpIDMin is the first storage driver snapshot operation.
	SdSnapOpIDMin = SdSnapshots

	// SdSnapOpIDMax is the last storage driver snapshot operation.
	SdSnapOpIDMax = SdSnapshotRemove
)

// SdSupported is a mask of all of the available storage driver op IDs.
const SdSupported = 0 |
	SdInstanceID |
	SdNextDevice |
	SdLocalDevices |
	SdLogin |
	SdNextDeviceInfo |
	SdType |
	SdInstanceIDInspect |
	SdVolumes |
	SdVolumeInspect |
	SdVolumeInspectByName |
	SdVolumeCreate |
	SdVolumeCreateFromSnapshot |
	SdVolumeCopy |
	SdVolumeSnapshot |
	SdVolumeRemove |
	SdVolumeAttach |
	SdVolumeDetach |
	SdSnapshots |
	SdSnapshotInspect |
	SdSnapshotCopy |
	SdSnapshotRemove

const (
	// StBlock is the block storage type.
	StBlock uint8 = 1 + iota

	// StNAS is the network attached storage type.
	StNAS

	// StObject is the object-backed storage type.
	StObject
)

// VolumesOpts are options when inspecting a volume.
type VolumesOpts interface {

	// GetAttachments returns the mask that indicates whether or not to
	// return a volume based on its attachment status.
	GetAttachments() uint64

	// GetOpts returns a Store.
	GetOpts() interface{}
}

// VolumeInspectOpts are options when inspecting a volume.
type VolumeInspectOpts interface {

	// GetAttachments returns the mask that indicates whether or not to
	// return a volume based on its attachment status.
	GetAttachments() uint64

	// GetOpts returns a Store.
	GetOpts() interface{}
}

// VolumeCreateOpts are options when creating a new volume.
type VolumeCreateOpts interface {

	// GetAvailabilityZone returns the availability zone of the new
	// volume.
	GetAvailabilityZone() *string

	// GetIOPS returns the requested IOPS for the new volume.
	GetIOPS() *int64

	// GetSize returns the size (in bytes) for the new volume.
	GetSize() *int64

	// GetType returns the type of the new volume.
	GetType() *string

	// GetEncrypted returns a flag that indicates whether or not to
	// encrypt a new volume.
	IsEncrypted() *bool

	// GetEncryptionKey returns the key used to encrypt a new volume.
	GetEncryptionKey() *string

	// GetOpts returns a Store.
	GetOpts() interface{}
}

// VolumeAttachOpts are options for attaching a volume.
type VolumeAttachOpts interface {

	// GetNextDevice returns the next device to use when attaching the
	// volume.
	GetNextDevice() *string

	// IsForced returns a flag indicating whether or not to preempt an
	// existing attachment in order to process this attachment request.
	IsForced() bool

	// GetOpts returns a Store.
	GetOpts() interface{}
}

// VolumeDetachOpts are options for detaching a volume.
type VolumeDetachOpts interface {

	// IsForced returns a flag indicating whether or not to detach the
	// volume regardless of its state.
	IsForced() bool

	// GetOpts returns a Store.
	GetOpts() interface{}
}

// VolumeRemoveOpts are options for removing a volume.
type VolumeRemoveOpts interface {

	// IsForced returns a flag indicating whether or not to remove the
	// volume regardless of its state.
	IsForced() bool

	// GetOpts returns a Store.
	GetOpts() interface{}
}

const (
	// VolumeAttachmentStateUnknown indicates the driver has set the state,
	// but it is explicitly unknown and should not be inferred from the list of
	// attachments alone.
	VolumeAttachmentStateUnknown uint8 = 1

	// VolumeAttached indicates the volume is attached to the instance
	// specified in the API call that requested the volume information.
	VolumeAttached uint8 = 2

	// VolumeAvailable indicates the volume is not attached to any instance.
	VolumeAvailable uint8 = 3

	// VolumeUnavailable indicates the volume is attached to some instance
	// other than the one specified in the API call that requested the
	// volume information.
	VolumeUnavailable uint8 = 4
)

// Volume provides information about a storage volume.
type Volume interface {

	// GetAttachments returns information about the instances to which the
	// volume is attached. Each element of the array is a VolumeAttachment.
	GetAttachments() []interface{}

	// GetAttachmentState indicates whether or not a volume is attached. A client
	// can surmise the same state stored in this field by inspecting a volume's
	// Attachments field, but this field provides the server a means of doing
	// that inspection and storing the result so the client does not have to do
	// so.
	GetAttachmentState() uint64

	// GetAvailabilityZone returns the availability zone for which the volume is
	// available.
	GetAvailabilityZone() string

	// IsEncrypted returns a flag indicating whether or not the volume is
	// encrypted.
	IsEncrypted() bool

	// GetIOPS returns the IOPS value for the volume.
	GetIOPS() int64

	// GetName returns the name of the volume.
	GetName() string

	// GetNetworkName returns the name the device is known by on the
	// system(s) to which the device is attached.
	GetNetworkName() string

	// GetSize returns the size of the volume in bytes.
	GetSize() int64

	// GetStatus returns the status of the volume.
	GetStatus() string

	// GetID returns an identifier unique to the storage platform to which
	// the volume belongs. A volume ID is not guaranteed to be unique
	// across multiple storage platforms.
	GetID() string

	// GetType returns type of storage the volume provides.
	GetType() uint8

	// GetFields returns additional information about the object.
	GetFields() map[string]string
}

// VolumeAttachment provides information about an object attached to a
// storage volume.
type VolumeAttachment interface {

	// The name of the device on which the volume to which the object is
	// attached is mounted.
	GetDeviceName() string

	// MountPoint is the mount point for the volume. This field is set when a
	// volume is retrieved via an integration driver.
	GetMountPoint() string

	// The ID of the instance on which the volume to which the attachment
	// belongs is mounted.
	GetInstanceID() interface{}

	// The status of the attachment.
	GetStatus() string

	// The ID of the volume to which the attachment belongs.
	GetVolumeID() string

	// GetFields returns additional information about the object.
	GetFields() map[string]string
}

// MountInfo reveals information about a particular mounted filesystem. This
// struct is populated from the content in the /proc/<pid>/mountinfo file.
type MountInfo interface {

	// ID is a unique identifier of the mount (may be reused after umount).
	GetID() int

	// Parent indicates the ID of the mount parent (or of self for the top of
	// the mount tree).
	GetParent() int

	// Major indicates one half of the device ID which identifies the device
	// class.
	GetMajor() int

	// Minor indicates one half of the device ID which identifies a specific
	// instance of device.
	GetMinor() int

	// Root of the mount within the filesystem.
	GetRoot() string

	// MountPoint indicates the mount point relative to the process's root.
	GetMountPoint() string

	// Opts represents mount-specific options.
	GetOpts() string

	// Optional represents optional fields.
	GetOptional() string

	// FSType indicates the type of filesystem, such as EXT3.
	GetFSType() string

	// Source indicates filesystem specific information or "none".
	GetSource() string

	// VFSOpts represents per super block options.
	GetVFSOpts() string

	// Fields returns additional information about the object.
	GetFields() map[string]string
}

// Snapshot provides information about a storage-layer snapshot.
type Snapshot interface {

	// A description of the snapshot.
	GetDescription() string

	// The name of the snapshot.
	GetName() string

	// A flag indicating whether or not the snapshot is encrypted.
	IsEncrypted() bool

	// The snapshot's ID.
	GetID() string

	// The time (epoch) at which the request to create the snapshot was submitted.
	GetStartTime() int64

	// The status of the snapshot.
	GetStatus() string

	// The ID of the volume to which the snapshot belongs.
	GetVolumeID() string

	// The size of the volume to which the snapshot belongs.
	GetVolumeSize() int64

	// Fields returns additional information about the object.
	GetFields() map[string]string
}

// NextDeviceInfo assists the libStorage client in determining the
// next available device name by providing the driver's device prefix and
// optional pattern.
//
// For example, the Amazon Web Services (AWS) device prefix is "xvd" and its
// pattern is "[a-z]". These two values would be used to determine on an EC2
// instance where "/dev/xvda" and "/dev/xvdb" are in use that the next
// available device name is "/dev/xvdc".
//
// If the Ignore field is set to true then the client logic does not invoke the
// GetNextAvailableDeviceName function prior to submitting an AttachVolume
// request to the server.
type NextDeviceInfo interface {
	// IsIgnored returns a flag that indicates whether the client logic
	// should invoke the GetNextAvailableDeviceName function prior to
	// submitting an AttachVolume request to the server.
	IsIgnored() bool

	// GetPrefix returns the first part of a device path's value after the
	// "/dev/" porition. For example, the prefix in "/dev/xvda" is "xvd".
	GetPrefix() string

	// GetPattern returns the regex to match the part of a device path after the
	// prefix.
	GetPattern() string

	// Fields returns additional information about the object.
	GetFields() map[string]string
}
