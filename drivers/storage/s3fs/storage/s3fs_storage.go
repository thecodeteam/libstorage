// +build !libstorage_storage_driver libstorage_storage_driver_s3fs

package storage

import (
	"fmt"
	"strings"

	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/goof"

	"github.com/codedellemc/libstorage/api/context"
	"github.com/codedellemc/libstorage/api/registry"
	"github.com/codedellemc/libstorage/api/types"

	"github.com/codedellemc/libstorage/drivers/storage/s3fs"
	"github.com/codedellemc/libstorage/drivers/storage/s3fs/utils"
)

type driver struct {
	name    string
	config  gofig.Config
	buckets []string
}

func init() {
	registry.RegisterStorageDriver(s3fs.Name, newDriver)
}

func newDriver() types.StorageDriver {
	return &driver{name: s3fs.Name}
}

func (d *driver) Name() string {
	return d.name
}

// Init initializes the driver.
func (d *driver) Init(context types.Context, config gofig.Config) error {
	d.config = config

	// TODO: add options
	d.buckets = d.getBuckets()

	context.Info(fmt.Sprintf(
		"s3fs storage driver initialized: %s", d.buckets))
	return nil
}

// NextDeviceInfo returns the information about the driver's next available
func (d *driver) NextDeviceInfo(
	ctx types.Context) (*types.NextDeviceInfo, error) {
	return nil, nil
}

// Type returns the type of storage the driver provides.
func (d *driver) Type(ctx types.Context) (types.StorageType, error) {
	return types.Object, nil
	//	return types.NAS, nil
}

// InstanceInspect returns an instance.
func (d *driver) InstanceInspect(
	ctx types.Context,
	opts types.Store) (*types.Instance, error) {

	iid := context.MustInstanceID(ctx)
	return &types.Instance{
		Name: iid.ID,
		// Region:       iid.Fields[s3fs.InstanceIDFieldRegion],
		InstanceID:   iid,
		ProviderName: iid.Driver,
	}, nil
}

// Volumes returns all volumes or a filtered list of volumes.
func (d *driver) Volumes(
	ctx types.Context,
	opts *types.VolumesOpts) ([]*types.Volume, error) {

	// Convert retrieved volumes to libStorage types.Volume
	vols, convErr := d.toTypesVolume(ctx, &d.buckets, opts.Attachments)
	if convErr != nil {
		return nil, goof.WithError(
			"error converting to types.Volume", convErr)
	}

	ctx.Debugf("DBG: volumes: %s", vols)

	return vols, nil
}

// VolumeInspect inspects a single volume.
func (d *driver) VolumeInspect(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeInspectOpts) (*types.Volume, error) {

	return d.getVolume(ctx, volumeID, opts.Attachments)
}

// VolumeCreate creates a new volume.
func (d *driver) VolumeCreate(ctx types.Context, volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {

	if opts.Encrypted != nil && *opts.Encrypted {
		return nil, types.ErrNotImplemented
	}

	// TODO: bucket creation is not supported by s3fs,
	// possible options to implement if needed (probably both could be
	// implemented and behaviour could be switched via config):
	//	- implement bucket creation via S3 REST API
	//	- use 'backet:/path' semantic, where bucket should exist
	//	  and '/path' is a volume description object, in that case
	//	  all volumes would be in one bucket.
	//        WARN: This options doesn't work in s3fs on my environemnt!!!
	// For now it just returns volume object if bucket is in available
	// list from config
	volume, err := d.getVolume(ctx, volumeName, types.VolAttNone)
	if err != nil {
		return nil, goof.WithError(
			"Volume is not in the list of allowed ones in config",
			err)
	}
	return volume, nil
}

// VolumeCreateFromSnapshot creates a new volume from an existing snapshot.
func (d *driver) VolumeCreateFromSnapshot(
	ctx types.Context,
	snapshotID, volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {
	// TODO Snapshots are not implemented yet
	return nil, types.ErrNotImplemented
}

// VolumeCopy copies an existing volume.
func (d *driver) VolumeCopy(
	ctx types.Context,
	volumeID, volumeName string,
	opts types.Store) (*types.Volume, error) {
	// TODO Snapshots are not implemented yet
	return nil, types.ErrNotImplemented
}

// VolumeSnapshot snapshots a volume.
func (d *driver) VolumeSnapshot(
	ctx types.Context,
	volumeID, snapshotName string,
	opts types.Store) (*types.Snapshot, error) {
	// TODO Snapshots are not implemented yet
	return nil, types.ErrNotImplemented
}

// VolumeRemove removes a volume.
func (d *driver) VolumeRemove(
	ctx types.Context,
	volumeID string,
	opts types.Store) error {

	fields := map[string]interface{}{
		"provider": d.Name(),
		"volumeID": volumeID,
	}
	_, err := d.getVolume(ctx, volumeID, types.VolAttNone)
	if err != nil {
		return goof.WithFields(fields, "volume does not exist")
	}

	// TODO: see comment in VolumeCreate,
	// For now there is nothing to do
	return nil
}

// VolumeAttach attaches a volume and provides a token clients can use
// to validate that device has appeared locally.
func (d *driver) VolumeAttach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeAttachOpts) (*types.Volume, string, error) {

	fields := map[string]interface{}{
		"provider": d.Name(),
		"volumeID": volumeID,
	}

	volume, err := d.getVolume(ctx, volumeID,
		types.VolumeAttachmentsRequested)
	if err != nil {
		return nil, "", goof.WithFieldsE(fields,
			"failed to get volume for attach", err)
	}

	// Nothing to do for attach
	return volume, "", nil
}

// VolumeDetach detaches a volume.
func (d *driver) VolumeDetach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeDetachOpts) (*types.Volume, error) {

	fields := map[string]interface{}{
		"provider": d.Name(),
		"volumeID": volumeID,
	}

	volume, err := d.getVolume(ctx, volumeID,
		types.VolumeAttachmentsRequested)
	if err != nil {
		return nil, goof.WithFieldsE(
			fields, "failed to get volume", err)
	}

	// Nothing to do for detach
	return volume, nil
}

// Snapshots returns all volumes or a filtered list of snapshots.
func (d *driver) Snapshots(
	ctx types.Context,
	opts types.Store) ([]*types.Snapshot, error) {
	// TODO Snapshots are not implemented yet
	return nil, types.ErrNotImplemented
}

// SnapshotInspect inspects a single snapshot.
func (d *driver) SnapshotInspect(
	ctx types.Context,
	snapshotID string,
	opts types.Store) (*types.Snapshot, error) {
	// TODO Snapshots are not implemented yet
	return nil, types.ErrNotImplemented
}

// SnapshotCopy copies an existing snapshot.
func (d *driver) SnapshotCopy(
	ctx types.Context,
	snapshotID, snapshotName, destinationID string,
	opts types.Store) (*types.Snapshot, error) {
	// TODO Snapshots are not implemented yet
	return nil, types.ErrNotImplemented
}

// SnapshotRemove removes a snapshot.
func (d *driver) SnapshotRemove(
	ctx types.Context,
	snapshotID string,
	opts types.Store) error {
	// TODO Snapshots are not implemented yet
	return types.ErrNotImplemented
}

// Retrieve config arguments
func (d *driver) getCredFilePath() string {
	return d.config.GetString(s3fs.ConfigS3FSCredFilePathKey)
}

func (d *driver) getBuckets() []string {
	result := d.config.GetString(s3fs.ConfigS3FSBucketsKey)
	return strings.Split(result, ",")
}

func (d *driver) getTag() string {
	return d.config.GetString(s3fs.ConfigS3FSTagKey)
}

var errGetLocDevs = goof.New("error getting local devices from context")

func (d *driver) toTypesVolume(
	ctx types.Context,
	buckets *[]string,
	attachments types.VolumeAttachmentsTypes) ([]*types.Volume, error) {

	var volumesSD []*types.Volume
	for _, bucket := range *buckets {
		volumeSD, err := d.toTypeVolume(ctx, bucket, attachments)
		if err != nil {
			return nil, goof.WithError(
				"Failed to convert volume", err)
		} else if volumeSD != nil {
			volumesSD = append(volumesSD, volumeSD)
		}
	}
	return volumesSD, nil
}

const (
	psOutputFormat = "%s %s %s %s %s"
)

// TODO:
//   - it should be run on client side...
//     so it is needed to do something with it..
//   - it is code duplication with os driver part
//     probably it useless here nad could be just removed
func (d *driver) getMountedBuckets(ctx types.Context) ([]string, error) {

	var bucketsMap map[string]string
	var err error
	if bucketsMap, err = utils.MountedBuckets(ctx); err != nil {
		return nil, err
	}
	var buckets []string
	for b := range bucketsMap {
		buckets = append(buckets, b)
	}
	ctx.Debug("DBG: mounted buckets: %s", buckets)
	return buckets, nil
}

func find(array []string, element string) bool {
	for _, item := range array {
		if item == element {
			return true
		}
	}
	return false
}

func (d *driver) toTypeVolume(
	ctx types.Context,
	bucket string,
	attachments types.VolumeAttachmentsTypes) (*types.Volume, error) {

	var buckets []string
	var err error
	if buckets, err = d.getMountedBuckets(ctx); err != nil {
		return nil, goof.WithError(fmt.Sprintf(
			"Failed to convert bucket '%s' to volume", bucket),
			err)
	}

	attachmentStatus := "Exported and Unmounted"
	if find(buckets, bucket) {
		attachmentStatus = "Exported and Mounted"
	}
	var attachmentsSD []*types.VolumeAttachment
	if attachments.Requested() {
		id, ok := context.InstanceID(ctx)
		if !ok || id == nil {
			return nil, goof.New("Can't get instance ID to filter volume")
		}
		attachmentsSD = append(
			attachmentsSD,
			&types.VolumeAttachment{
				InstanceID: id,
				VolumeID:   bucket,
				DeviceName: utils.BucketURI(bucket),
				Status:     attachmentStatus,
			})
	}

	volumeSD := &types.Volume{
		Name:        bucket,
		ID:          bucket,
		Attachments: attachmentsSD,
		// TODO:
		//AvailabilityZone: *volume.AvailabilityZone,
		//Encrypted:        *volume.Encrypted,
	}

	// Some volume types have no IOPS, so we get nil in volume.Iops
	//if volume.Iops != nil {
	//	volumeSD.IOPS = *volume.Iops
	//}

	return volumeSD, nil
}

func (d *driver) getVolume(
	ctx types.Context,
	volumeID string,
	attachments types.VolumeAttachmentsTypes) (*types.Volume, error) {

	for _, bucket := range d.buckets {
		if bucket == volumeID {
			return d.toTypeVolume(ctx, bucket, attachments)
		}
	}
	return nil, fmt.Errorf("Error to get volume %s", volumeID)
}
