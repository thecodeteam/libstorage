package executor

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/akutz/gofig"
	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
	"github.com/emccode/libstorage/api/types/drivers"
)

const Name = "scaleio"

// Executor is the storage executor for the ScaleIO storage driver.
type StorageExecutor struct {
	Config     gofig.Config
	name       string
	instanceID *types.InstanceID
	InitDriver func() error
}

func init() {
	gofig.Register(configRegistration())
	registry.RegisterStorageExecutor(Name, newExecutor)
}

func (e *StorageExecutor) Init(config gofig.Config) error {
	e.Config = config
	if e.InitDriver != nil {
		if err := e.InitDriver(); err != nil {
			return err
		}
	}
	instanceID, err := getSdcLocalGUID()
	if err != nil {
		return err
	}
	e.instanceID = &types.InstanceID{ID: instanceID}
	return nil
}

func newExecutor() drivers.StorageExecutor {
	return NewExecutor()
}

func NewExecutor() *StorageExecutor {
	return &StorageExecutor{
		name:       Name,
	}
}

func (e *StorageExecutor) Name() string {
	return e.name
}

// InstanceID returns the local system's InstanceID.
func (e *StorageExecutor) InstanceID(
	ctx context.Context,
	opts types.Store) (*types.InstanceID, error) {
	return e.instanceID, nil
}

// NextDevice returns the next available device.
func (e *StorageExecutor) NextDevice(
	ctx context.Context,
	opts types.Store) (string, error) {
	return "", nil
}

// LocalDevices returns a map of the system's local devices.
func (e *StorageExecutor) LocalDevices(
	ctx context.Context,
	opts types.Store) (map[string]string, error) {

	var volumeMap = make(map[string]string)

	diskIDPath := "/dev/disk/by-id"
	files, _ := ioutil.ReadDir(diskIDPath)
	r, _ := regexp.Compile(`^emc-vol-\w*-\w*$`)
	for _, f := range files {
		matched := r.MatchString(f.Name())
		if matched {
			mdmVolumeID := strings.Replace(f.Name(), "emc-vol-", "", 1)
			devPath, _ := filepath.EvalSymlinks(fmt.Sprintf("%s/%s", diskIDPath, f.Name()))
			volumeID := strings.Split(mdmVolumeID, "-")[1]
			volumeMap[volumeID] = devPath
		}
	}
	return volumeMap, nil
}

func configRegistration() *gofig.Registration {
	r := gofig.NewRegistration("ScaleIO")
	r.Key(gofig.String, "", "", "", "scaleio.endpoint")
	r.Key(gofig.Bool, "", false, "", "scaleio.insecure")
	r.Key(gofig.Bool, "", false, "", "scaleio.useCerts")
	r.Key(gofig.String, "", "", "", "scaleio.userID")
	r.Key(gofig.String, "", "", "", "scaleio.userName")
	r.Key(gofig.String, "", "", "", "scaleio.password")
	r.Key(gofig.String, "", "", "", "scaleio.systemID")
	r.Key(gofig.String, "", "", "", "scaleio.systemName")
	r.Key(gofig.String, "", "", "", "scaleio.protectionDomainID")
	r.Key(gofig.String, "", "", "", "scaleio.protectionDomainName")
	r.Key(gofig.String, "", "", "", "scaleio.storagePoolID")
	r.Key(gofig.String, "", "", "", "scaleio.storagePoolName")
	r.Key(gofig.String, "", "", "", "scaleio.thinOrThick")
	r.Key(gofig.String, "", "", "", "scaleio.version")
	return r
}

func getSdcLocalGUID() (sdcGUID string, err error) {

	// get sdc kernel guid
	// /bin/emc/scaleio/drv_cfg --query_guid
	// sdcKernelGuid := "271bad82-08ee-44f2-a2b1-7e2787c27be1"

	out, err := exec.Command("/opt/emc/scaleio/sdc/bin/drv_cfg", "--query_guid").Output()
	if err != nil {
		return "", fmt.Errorf("Error querying volumes: ", err)
	}

	sdcGUID = strings.Replace(string(out), "\n", "", -1)

	return sdcGUID, nil
}


