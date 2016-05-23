// +build linux darwin

/*
Package unix is the OS driver for linux and darwin. In order to reduce external
dependencies, this package borrows the following packages:

  - github.com/docker/docker/pkg/mount
  - github.com/opencontainers/runc/libcontainer/label
*/
package unix

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/akutz/gofig"
	"github.com/akutz/goof"

	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
)

var driverName = runtime.GOOS

var (
	errUnknownFileSystem     = goof.New("unknown file system")
	errUnsupportedFileSystem = goof.New("unsupported file system")
)

func init() {
	registry.RegisterOSDriver(driverName, newDriver)
	gofig.Register(configRegistration())
}

type driver struct {
	config gofig.Config
}

func newDriver() types.OSDriver {
	return &driver{}
}

func (d *driver) Init(ctx types.Context, config gofig.Config) error {
	d.config = config
	return nil
}

func (d *driver) Name() string {
	return driverName
}

func (d *driver) Mounts(
	ctx types.Context,
	deviceName, mountPoint string,
	opts types.Store) ([]*types.MountInfo, error) {

	mounts, err := mounts(ctx, deviceName, mountPoint, opts)
	if err != nil {
		return nil, err
	}

	if mountPoint == "" && deviceName == "" {
		return mounts, nil
	} else if mountPoint != "" && deviceName != "" {
		return nil, goof.New("cannot specify mountPoint and deviceName")
	}

	matchedMounts := []*types.MountInfo{}
	for _, m := range mounts {
		if m.MountPoint == mountPoint || m.Source == deviceName {
			matchedMounts = append(matchedMounts, m)
		}
	}
	return matchedMounts, nil
}

func (d *driver) Mount(
	ctx types.Context,
	deviceName, mountPoint string,
	opts *types.DeviceMountOpts) error {

	if d.isNfsDevice(deviceName) {

		if err := d.nfsMount(deviceName, mountPoint); err != nil {
			return err
		}

		os.MkdirAll(d.volumeMountPath(mountPoint), d.fileModeMountPath())
		os.Chmod(d.volumeMountPath(mountPoint), d.fileModeMountPath())

		return nil
	}

	fsType, err := probeFsType(deviceName)
	if err != nil {
		return err
	}

	options := formatMountLabel("", opts.MountLabel)
	options = fmt.Sprintf("%s,%s", opts.MountOptions, opts.MountLabel)
	if fsType == "xfs" {
		options = fmt.Sprintf("%s,nouuid", opts.MountLabel)
	}

	if err := mount(deviceName, mountPoint, fsType, options); err != nil {
		return goof.WithFieldsE(goof.Fields{
			"deviceName": deviceName,
			"mountPoint": mountPoint,
		}, "error mounting directory", err)
	}

	os.MkdirAll(d.volumeMountPath(mountPoint), d.fileModeMountPath())
	os.Chmod(d.volumeMountPath(mountPoint), d.fileModeMountPath())

	return nil
}

func (d *driver) Unmount(
	ctx types.Context,
	mountPoint string,
	opts types.Store) error {

	var (
		err       error
		isMounted bool
	)

	isMounted, err = d.IsMounted(ctx, mountPoint, opts)
	if err != nil || !isMounted {
		return err
	}

	for i := 0; i < 10; i++ {
		if err = syscall.Unmount(mountPoint, 0); err == nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

func (d *driver) IsMounted(
	ctx types.Context,
	mountPoint string,
	opts types.Store) (bool, error) {

	entries, err := mounts(ctx, "", mountPoint, opts)
	if err != nil {
		return false, err
	}

	// Search the table for the mountpoint
	for _, e := range entries {
		if e.MountPoint == mountPoint {
			return true, nil
		}
	}
	return false, nil
}

func (d *driver) Format(
	ctx types.Context,
	deviceName string,
	opts *types.DeviceFormatOpts) error {

	return format(ctx, deviceName, opts)
}

func (d *driver) isNfsDevice(device string) bool {
	return strings.Contains(device, ":")
}

func (d *driver) nfsMount(device, target string) error {

	command := exec.Command("mount", device, target)
	output, err := command.CombinedOutput()
	if err != nil {
		return goof.WithError(fmt.Sprintf("failed mounting: %s", output), err)
	}

	return nil
}

func (d *driver) fileModeMountPath() (fileMode os.FileMode) {
	return os.FileMode(d.volumeFileMode())
}

// from github.com/docker/docker/daemon/graphdriver/devmapper/
// this should be abstracted outside of graphdriver but within Docker package,
// here temporarily
type probeData struct {
	fsName string
	magic  string
	offset uint64
}

func probeFsType(device string) (string, error) {
	probes := []probeData{
		{"btrfs", "_BHRfS_M", 0x10040},
		{"ext4", "\123\357", 0x438},
		{"xfs", "XFSB", 0},
	}

	maxLen := uint64(0)
	for _, p := range probes {
		l := p.offset + uint64(len(p.magic))
		if l > maxLen {
			maxLen = l
		}
	}

	file, err := os.Open(device)
	if err != nil {
		return "", err
	}
	defer file.Close()

	buffer := make([]byte, maxLen)
	l, err := file.Read(buffer)
	if err != nil {
		return "", err
	}

	if uint64(l) != maxLen {
		return "", goof.WithField(
			"device", device, "error detecting filesystem")
	}

	for _, p := range probes {
		if bytes.Equal(
			[]byte(p.magic), buffer[p.offset:p.offset+uint64(len(p.magic))]) {
			return p.fsName, nil
		}
	}

	return "", errUnknownFileSystem
}

/*
formatMountLabel returns a string to be used by the mount command.
The format of this string will be used to alter the labeling of the mountpoint.
The string returned is suitable to be used as the options field of the mount
command.

If you need to have additional mount point options, you can pass them in as
the first parameter.  Second parameter is the label that you wish to apply
to all content in the mount point.
*/
func formatMountLabel(src, mountLabel string) string {
	if mountLabel != "" {
		switch src {
		case "":
			src = fmt.Sprintf("context=%q", mountLabel)
		default:
			src = fmt.Sprintf("%s,context=%q", src, mountLabel)
		}
	}
	return src
}

func (d *driver) volumeMountPath(target string) string {
	return fmt.Sprintf("%s%s", target, d.volumeRootPath())
}

func (d *driver) volumeFileMode() int {
	return d.config.GetInt("linux.volume.filemode")
}

func (d *driver) volumeRootPath() string {
	return d.config.GetString("linux.volume.rootpath")
}

func configRegistration() *gofig.Registration {
	r := gofig.NewRegistration("Linux")
	r.Key(gofig.Int, "", 0700, "", "linux.volume.filemode")
	r.Key(gofig.String, "", "/data", "", "linux.volume.rootpath")
	return r
}
