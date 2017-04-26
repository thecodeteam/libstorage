// +build linux darwin
// +build !libstorage_storage_executor libstorage_storage_executor_vmdk

package executor

import (
	gofig "github.com/akutz/gofig/types"

	"github.com/codedellemc/libstorage/api/registry"
	"github.com/codedellemc/libstorage/api/types"

	"github.com/codedellemc/libstorage/drivers/storage/vmdk"
)

type driver struct {
	config gofig.Config
}

func init() {
	registry.RegisterStorageExecutor(vmdk.Name, newDriver)
}

func newDriver() types.StorageExecutor {
	return &driver{}
}

func (d *driver) Name() string {
	return vmdk.Name
}

func (d *driver) Supported(
	ctx types.Context,
	opts types.Store) (bool, error) {

	return true, nil
}

func (d *driver) Init(ctx types.Context, config gofig.Config) error {
	d.config = config
	return nil
}

// InstanceID returns the local system's InstanceID.
func (d *driver) InstanceID(
	ctx types.Context,
	opts types.Store) (*types.InstanceID, error) {

	iid := &types.InstanceID{Driver: vmdk.Name}
	iid.ID = vmdk.Name
	return iid, nil
}

// NextDevice returns the next available device.
func (d *driver) NextDevice(
	ctx types.Context,
	opts types.Store) (string, error) {

	return "", nil
}

// LocalDevices returns a map of the system's local devices.
func (d *driver) LocalDevices(
	ctx types.Context,
	opts *types.LocalDevicesOpts) (*types.LocalDevices, error) {

	return &types.LocalDevices{DeviceMap: map[string]string{}}, nil
}
