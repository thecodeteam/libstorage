// +build !libstorage_storage_driver libstorage_storage_driver_vmdk

package storage

import (
	"fmt"
	"regexp"
	"strconv"
	"sync"

	gofig "github.com/akutz/gofig/types"

	"github.com/codedellemc/libstorage/api/registry"
	"github.com/codedellemc/libstorage/api/types"

	"github.com/codedellemc/libstorage/drivers/storage/vmdk"
	"github.com/vmware/docker-volume-vsphere/vmdk_plugin/drivers/vmdk/vmdkops"
)

const (
	minSizeGiB = 1
)

type driver struct {
	ctx    types.Context
	config gofig.Config
	ops    vmdkops.VmdkOps
}

func init() {
	registry.RegisterStorageDriver(vmdk.Name, newDriver)
}

func newDriver() types.StorageDriver {
	return &driver{}
}

func (d *driver) Name() string {
	return vmdk.Name
}

func (d *driver) Type(ctx types.Context) (types.StorageType, error) {
	return types.Block, nil
}

func (d *driver) Init(ctx types.Context, config gofig.Config) error {
	d.ctx = ctx
	d.config = config
	d.ops = vmdkops.VmdkOps{Cmd: vmdkops.EsxVmdkCmd{Mtx: &sync.Mutex{}}}
	return nil
}

func (d *driver) NextDeviceInfo(
	ctx types.Context) (*types.NextDeviceInfo, error) {
	return &types.NextDeviceInfo{
		Ignore: true,
	}, nil
}

func (d *driver) InstanceInspect(
	ctx types.Context,
	opts types.Store) (*types.Instance, error) {

	return nil, nil
}

func (d *driver) Volumes(
	ctx types.Context,
	opts *types.VolumesOpts) ([]*types.Volume, error) {

	data, err := d.ops.List()
	if err != nil {
		return nil, err
	}

	vols := []*types.Volume{}
	for _, v := range data {
		vols = append(vols, &types.Volume{
			Name:   v.Name,
			Fields: v.Attributes,
		})
	}

	return vols, nil
}

func (d *driver) VolumeInspect(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeInspectOpts) (*types.Volume, error) {

	data, err := d.ops.Get(volumeID)
	if err != nil {
		return nil, err
	}

	vol := &types.Volume{
		ID:     volumeID,
		Name:   volumeID,
		Fields: map[string]string{},
	}

	for k, v := range data {
		switch k {
		case "datastore":
			if v, ok := v.(string); ok {
				vol.Type = v
			}
		case "capacity":
			if v, ok := v.(map[string]interface{}); ok {
				if v, ok := v["size"].(string); ok {
					if v == "0" {
						vol.Size = 0
					} else if sz, ok := isGB(v); ok {
						vol.Size = sz * 1024 * 1024 * 1024
					} else if sz, ok := isMB(v); ok {
						vol.Size = sz * 1024 * 1024
					} else if sz, ok := isKB(v); ok {
						vol.Size = sz * 1024
					}
				}
			}
		default:
			vol.Fields[k] = fmt.Sprintf("%v", v)
		}
	}

	return vol, nil
}

func isSize(rx *regexp.Regexp, s string) (int64, bool) {
	m := rx.FindStringSubmatch(s)
	if len(m) == 0 {
		return 0, false
	}
	i, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, false
	}
	return int64(i), true
}

var rxIsKB = regexp.MustCompile(`(?i)^([\d,\.]+)\s*KB\s*$`)

func isKB(s string) (int64, bool) {
	return isSize(rxIsKB, s)
}

var rxIsMB = regexp.MustCompile(`(?i)^([\d,\.]+)\s*MB\s*$`)

func isMB(s string) (int64, bool) {
	return isSize(rxIsMB, s)
}

var rxIsGB = regexp.MustCompile(`(?i)^([\d,\.]+)\s*GB\s*$`)

func isGB(s string) (int64, bool) {
	return isSize(rxIsGB, s)
}

func (d *driver) VolumeCreate(
	ctx types.Context,
	name string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {

	if opts.Type != nil && *opts.Type != "" {
		name = fmt.Sprintf("%s@%s", name, *opts.Type)
	}

	createOpts := map[string]string{}
	if opts.Size != nil {
		createOpts["size"] = fmt.Sprintf("%dgb", *opts.Size)
	}

	if err := d.ops.Create(name, createOpts); err != nil {
		return nil, err
	}

	return d.VolumeInspect(ctx, name, nil)
}

func (d *driver) VolumeCreateFromSnapshot(
	ctx types.Context,
	snapshotID, volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {

	return nil, types.ErrNotImplemented
}

func (d *driver) VolumeCopy(
	ctx types.Context,
	volumeID, volumeName string,
	opts types.Store) (*types.Volume, error) {

	return nil, types.ErrNotImplemented
}

func (d *driver) VolumeSnapshot(
	ctx types.Context,
	volumeID, snapshotName string,
	opts types.Store) (*types.Snapshot, error) {

	return nil, types.ErrNotImplemented
}

func (d *driver) VolumeRemove(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeRemoveOpts) error {

	return d.ops.Remove(volumeID, map[string]string{})
}

func (d *driver) VolumeAttach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeAttachOpts) (*types.Volume, string, error) {

	dev, err := d.ops.Attach(volumeID, nil)
	if err != nil {
		return nil, "", err
	}

	var vol *types.Volume
	if vol, err = d.VolumeInspect(ctx, volumeID, nil); err != nil {
		return nil, "", err
	}

	return vol, string(dev), nil
}

func (d *driver) VolumeDetach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeDetachOpts) (*types.Volume, error) {

	if err := d.ops.Detach(volumeID, nil); err != nil {
		return nil, err
	}

	vol, err := d.VolumeInspect(ctx, volumeID, nil)
	if err != nil {
		return nil, err
	}

	return vol, nil
}

func (d *driver) Snapshots(
	ctx types.Context,
	opts types.Store) ([]*types.Snapshot, error) {

	return nil, types.ErrNotImplemented
}

func (d *driver) SnapshotInspect(
	ctx types.Context,
	snapshotID string,
	opts types.Store) (*types.Snapshot, error) {

	return nil, types.ErrNotImplemented
}

func (d *driver) SnapshotCopy(
	ctx types.Context,
	snapshotID, snapshotName, destinationID string,
	opts types.Store) (*types.Snapshot, error) {

	return nil, types.ErrNotImplemented
}

func (d *driver) SnapshotRemove(
	ctx types.Context,
	snapshotID string,
	opts types.Store) error {

	return types.ErrNotImplemented
}
