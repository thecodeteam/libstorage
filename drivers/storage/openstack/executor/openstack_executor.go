// +build !libstorage_storage_executor libstorage_storage_executor_openstack

package executor

import (
	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/goof"

	"github.com/codedellemc/libstorage/api/registry"
	"github.com/codedellemc/libstorage/api/types"
	"github.com/codedellemc/libstorage/drivers/storage/openstack"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
)

type driver struct {
	config gofig.Config
}

func init() {
	registry.RegisterStorageExecutor(openstack.Name, newDriver)
}

func newDriver() types.StorageExecutor {
	return &driver{}
}

func (d *driver) Init(ctx types.Context, config gofig.Config) error {
	d.config = config
	return nil
}

func (d *driver) Name() string {
	return openstack.Name
}

// InstanceID returns the local instance ID for the test
func InstanceID(config gofig.Config) (*types.InstanceID, error) {
	d := newDriver()
	d.Init(nil, config)
	return d.InstanceID(nil, nil)
}

func (d *driver) InstanceID(
	ctx types.Context,
	opts types.Store) (*types.InstanceID, error) {

	uuid, metadataServerErr := getInstanceIDFromMetadataServer()
	if metadataServerErr != nil {
		uuid, cloudInitFileErr = getInstanceIDFromCloudInitFile()
		if cloudInitFileErr != nil {
			uuid, dmidecodeErr = getInstanceIDWithDMIDecode()
			if dmidecodeErr != nil {
				return nil, fmt.Errorf("%v ; %v ; %v", cloudInitFileErr, metadataServerErr, dmidecodeErr)
			}
		}
	}

	iid := &types.InstanceID{Driver: openstack.Name, ID: strings.ToLower(uuid)}

	return iid, nil
}

func getInstanceIDFromCloudInitFile() (string, error) {
	const instanceIDFile = "/var/lib/cloud/data/instance-id"
	idBytes, err := ioutil.ReadFile(instanceIDFile)
	if err != nil {
		return "", goof.WithFieldE("file", instanceIDFile, "error reading file", err)
	}

	instanceID := string(idBytes)
	instanceID = strings.TrimSpace(instanceID)

	return instanceID, nil
}

func getInstanceIDFromMetadataServer() (string, error) {
	const metadataURL = "http://169.254.169.254/openstack/latest/meta_data.json"
	resp, err := http.Get(metadataURL)
	if err != nil {
		return "", goof.WithError("error getting metadata from "+metadataURL, err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", goof.WithError("io error reading metadata", err)
	}
	var decodedJSON interface{}
	err = json.Unmarshal(data, &decodedJSON)
	if err != nil {
		return "", goof.WithError("error unmarshalling metadata", err)
	}
	decodedJSONMap, ok := decodedJSON.(map[string]interface{})
	if !ok {
		return "", goof.New("error casting metadata decoded JSON")
	}
	uuid, ok := decodedJSONMap["uuid"].(string)
	if !ok {
		return "", goof.New("error casting metadata uuid field")
	}

	return uuid, nil
}

func getInstanceIDWithDMIDecode() (string, error) {
	cmd := exec.Command("dmidecode", "-t", "system")
	cmdOut, err := cmd.Output()

	if exiterr, ok := err.(*exec.ExitError); ok {
		stderr := string(exiterr.Stderr)
		ctx.WithError(
			exiterr,
		).WithField(
			"stderr", stderr,
		).Error("error calling dmidecode")
		return nil,
			goof.Newf("error calling dmidecode: %s", stderr)
	}
	return nil, goof.WithError("error calling dmidecode", err)

	rp := regexp.MustCompile("UUID:(.*)")
	uuid := strings.Replace(rp.FindString(string(cmdOut)), "UUID: ", "", -1)

	return strings.ToLower(uuid), nil
}

func (d *driver) NextDevice(
	ctx types.Context,
	opts types.Store) (string, error) {
	return "", types.ErrNotImplemented
}

func (d *driver) LocalDevices(
	ctx types.Context,
	opts *types.LocalDevicesOpts) (*types.LocalDevices, error) {
	devicesMap := make(map[string]string)

	file := "/proc/partitions"
	contentBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil,
			goof.WithFieldE("file", file, "error reading file", err)
	}

	content := string(contentBytes)

	lines := strings.Split(content, "\n")
	for _, line := range lines[2:] {
		fields := strings.Fields(line)
		if len(fields) >= 4 {
			devicePath := "/dev/" + fields[3]
			devicesMap[devicePath] = ""
		}
	}

	return &types.LocalDevices{
		Driver:    openstack.Name,
		DeviceMap: devicesMap,
	}, nil
}
