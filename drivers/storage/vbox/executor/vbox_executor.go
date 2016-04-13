package executor

import (
	"github.com/akutz/gofig"

	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
	"github.com/emccode/libstorage/api/types/drivers"
)

const (
	// Name is the name of the storage executor and driver.
	Name = "vbox"
)

// Executor is the storage executor for the VFS storage driver.
type Executor struct {

	// Config is the executor's configuration instance.
	Config gofig.Config
}

func init() {
	gofig.Register(configRegistration())

	registry.RegisterStorageExecutor(Name, newExecutor)
}

func newExecutor() drivers.StorageExecutor {
	return &Executor{}
}

func (d *Executor) Init(config gofig.Config) error {
	d.Config = config
	return nil
}

func (d *Executor) Name() string {
	return Name
}

// InstanceID returns the local system's InstanceID.
func (d *Executor) InstanceID(
	ctx context.Context,
	opts types.Store) (*types.InstanceID, error) {

	return nil, nil
}

// NextDevice returns the next available device.
func (d *Executor) NextDevice(
	ctx context.Context,
	opts types.Store) (string, error) {
	return "", nil
}

// LocalDevices returns a map of the system's local devices.
func (d *Executor) LocalDevices(
	ctx context.Context,
	opts types.Store) (map[string]string, error) {

	return nil, nil
}

// RootDir returns the path to the VFS root directory.
func (d *Executor) RootDir() string {
	return d.Config.GetString("vfs.root")
}

func configRegistration() *gofig.Registration {
	r := gofig.NewRegistration("virtualbox")
	r.Key(gofig.String, "", "", "", "virtualbox.endpoint")
	r.Key(gofig.String, "", "", "", "virtualbox.volumePath")
	r.Key(gofig.String, "", "", "", "virtualbox.localMachineNameOrId")
	r.Key(gofig.String, "", "", "", "virtualbox.username")
	r.Key(gofig.String, "", "", "", "virtualbox.password")
	r.Key(gofig.Bool, "", false, "", "virtualbox.tls")
	r.Key(gofig.String, "", "", "", "virtualbox.controllerName")
	return r
}
