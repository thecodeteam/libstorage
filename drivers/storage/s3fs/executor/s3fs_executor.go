// +build !libstorage_storage_executor libstorage_storage_executor_s3fs

package executor

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/goof"

	"github.com/codedellemc/libstorage/api/registry"
	"github.com/codedellemc/libstorage/api/types"

	"github.com/codedellemc/libstorage/drivers/storage/s3fs"
	"github.com/codedellemc/libstorage/drivers/storage/s3fs/utils"
)

const (
	// Template for parsing mount info file (/proc/self/mountinfo)
	mountinfoFormat = "%d %d %d:%d %s %s %s %s"
)

// driver is the storage executor for the s3fs storage driver.
type driver struct {
	name     string
	config   gofig.Config
	credFile string
}

func init() {
	registry.RegisterStorageExecutor(s3fs.Name, newDriver)
}

func newDriver() types.StorageExecutor {
	return &driver{name: s3fs.Name}
}

func (d *driver) Init(ctx types.Context, config gofig.Config) error {
	ctx.Info("s3fs_executor: Init")
	d.config = config
	if d.credFile = d.getCredFilePath(); d.credFile == "" {
		return goof.New(fmt.Sprintf(
			"%s mount driver requires %s option",
			d.name, s3fs.ConfigS3FSCredFilePathKey))
	}
	return nil
}

func (d *driver) Name() string {
	return d.name
}

// Supported returns a flag indicating whether or not the platform
// implementing the executor is valid for the host on which the executor
// resides.
func (d *driver) Supported(
	ctx types.Context,
	opts types.Store) (types.LSXSupportedOp, error) {

	supportedOp := types.LSXSOpNone
	var supp bool
	var err error
	if supp, err = utils.Supported(ctx); err != nil {
		return supportedOp, err
	}
	if supp {
		supportedOp = types.LSXSOpInstanceID |
			types.LSXSOpLocalDevices |
			types.LSXSOpMount
	}
	return supportedOp, nil
}

// InstanceID
func (d *driver) InstanceID(
	ctx types.Context,
	opts types.Store) (*types.InstanceID, error) {
	return utils.InstanceID(ctx)
}

// NextDevice returns the next available device.
func (d *driver) NextDevice(
	ctx types.Context,
	opts types.Store) (string, error) {
	return "", types.ErrNotImplemented
}

// Return list of local devices
func (d *driver) LocalDevices(
	ctx types.Context,
	opts *types.LocalDevicesOpts) (*types.LocalDevices, error) {

	mtt, err := parseMountTable(ctx)
	if err != nil {
		return nil, err
	}

	idmnt := make(map[string]string)
	for _, mt := range mtt {
		idmnt[mt.Source] = mt.MountPoint
	}

	return &types.LocalDevices{
		Driver:    d.name,
		DeviceMap: idmnt,
	}, nil
}

// Mount mounts a device to a specified path.
func (d *driver) Mount(
	ctx types.Context,
	deviceName, mountPoint string,
	opts *types.DeviceMountOpts) error {

	if !utils.IsS3FSURI(deviceName) {
		return goof.WithField(
			"device name", deviceName,
			"Unsupported device name format")
	}
	bucket := utils.BucketFromURI(deviceName)
	if mp, ok := utils.FindMountPoint(ctx, bucket); ok {
		ctx.Debugf("DBG: bucket '%s' is already mounted to '%s'",
			bucket, mp)
		if mp == mountPoint {
			// bucket is mounted to the required target => ok
			return nil
		}
		// bucket is mounted to another target => error
		return goof.WithFields(goof.Fields{
			"bucket":      bucket,
			"mount point": mp,
		}, "bucket is already mounted")
	}
	return utils.Mount(ctx, d.credFile, bucket, mountPoint, opts)
}

// Unmount unmounts the underlying device from the specified path.
func (d *driver) Unmount(
	ctx types.Context,
	mountPoint string,
	opts types.Store) error {

	return types.ErrNotImplemented
}

func (d *driver) getCredFilePath() string {
	return d.config.GetString(s3fs.ConfigS3FSCredFilePathKey)
}

func parseMountTable(ctx types.Context) ([]*types.MountInfo, error) {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return parseInfoFile(ctx, f)
}

func parseInfoFile(
	ctx types.Context,
	r io.Reader) ([]*types.MountInfo, error) {

	var (
		s   = bufio.NewScanner(r)
		out = []*types.MountInfo{}
	)

	for s.Scan() {
		if err := s.Err(); err != nil {
			return nil, err
		}

		var (
			p              = &types.MountInfo{}
			text           = s.Text()
			optionalFields string
		)

		if _, err := fmt.Sscanf(text, mountinfoFormat,
			&p.ID, &p.Parent, &p.Major, &p.Minor,
			&p.Root, &p.MountPoint, &p.Opts,
			&optionalFields); err != nil {

			return nil, fmt.Errorf("Scanning '%s' failed: %s",
				text, err)
		}
		// Safe as mountinfo encodes mountpoints with spaces as \040.
		index := strings.Index(text, " - ")
		postSeparatorFields := strings.Fields(text[index+3:])
		if len(postSeparatorFields) < 3 {
			return nil, fmt.Errorf(
				"Error found less than 3 fields post '-' in %q",
				text)
		}

		if optionalFields != "-" {
			p.Optional = optionalFields
		}

		p.FSType = postSeparatorFields[0]
		p.Source = postSeparatorFields[1]
		// s3fs doesnt provide mounted bucket, source is just 's3fs'
		// it is workaround - find bucket by mount point
		if strings.EqualFold(p.Source, s3fs.CmdName) {
			patchMountInfo(ctx, p)
		}
		p.VFSOpts = strings.Join(postSeparatorFields[2:], " ")
		out = append(out, p)
	}
	return out, nil
}

func patchMountInfo(ctx types.Context, m *types.MountInfo) {
	if m != nil && m.MountPoint != "" {
		if bucket, ok := utils.FindBucket(ctx, m.MountPoint); ok {
			m.Source = utils.BucketURI(bucket)
		}
	}
}
