package scaleio

import (
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/goof"

	"github.com/emccode/goscaleio"
	goscaleioTypes "github.com/emccode/goscaleio/types/v1"

	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/drivers/storage/scaleio/executor"
)

const (
	// Name is the name of the driver.
	Name = executor.Name
	cc   = 31
)

type driver struct {
	executor.StorageExecutor
	nextDeviceInfo   *types.NextDeviceInfo
	volumes          []*types.Volume
	snapshots        []*types.Snapshot
	client           *goscaleio.Client
	system           *goscaleio.System
	protectionDomain *goscaleio.ProtectionDomain
	storagePool      *goscaleio.StoragePool
	sdc              *goscaleio.Sdc
}

func init() {
	registry.RegisterStorageDriver(Name, newDriver)
}

func newDriver() types.StorageDriver {
	d := &driver{StorageExecutor: *executor.NewExecutor()}
	d.StorageExecutor.InitDriver = d.driverInit
	return d
}

func (d *driver) driverInit() error {

	fields := eff(map[string]interface{}{
		"endpoint": d.endpoint(),
		"version":  d.version(),
		"insecure": d.insecure(),
		"useCerts": d.useCerts(),
	})

	var err error

	if d.client, err = goscaleio.NewClientWithArgs(
		d.endpoint(),
		d.version(),
		d.insecure(),
		d.useCerts()); err != nil {
		return goof.WithFieldsE(fields, "error constructing new client", err)
	}
	if _, err := d.client.Authenticate(

		&goscaleio.ConfigConnect{
			d.endpoint(),
			d.version(),
			d.userName(),
			d.password()}); err != nil {
		fields["userName"] = d.userName()
		if d.password() != "" {
			fields["password"] = "******"
		}

		return goof.WithFieldsE(fields, "error authenticating", err)
	}
	if d.system, err = d.client.FindSystem(
		d.systemID(),
		d.systemName(), ""); err != nil {
		fields["systemId"] = d.systemID()
		fields["systemName"] = d.systemName()
		return goof.WithFieldsE(fields, "error finding system", err)
	}

	var pd *goscaleioTypes.ProtectionDomain
	if pd, err = d.system.FindProtectionDomain(
		d.protectionDomainID(),
		d.protectionDomainName(), ""); err != nil {
		fields["domainId"] = d.protectionDomainID()
		fields["domainName"] = d.protectionDomainName()
		return goof.WithFieldsE(fields,
			"error finding protection domain", err)
	}
	d.protectionDomain = goscaleio.NewProtectionDomain(d.client)
	d.protectionDomain.ProtectionDomain = pd

	var sp *goscaleioTypes.StoragePool
	if sp, err = d.protectionDomain.FindStoragePool(
		d.storagePoolID(),
		d.storagePoolName(), ""); err != nil {
		fields["storagePoolId"] = d.storagePoolID()
		fields["storagePoolName"] = d.storagePoolName()
		return goof.WithFieldsE(fields, "error finding storage pool", err)
	}
	d.storagePool = goscaleio.NewStoragePool(d.client)
	d.storagePool.StoragePool = sp

	var sdcGUID string
	if sdcGUID, err = goscaleio.GetSdcLocalGUID(); err != nil {
		return goof.WithFieldsE(fields, "error getting sdc local guid", err)
	}

	if d.sdc, err = d.system.FindSdc(
		"SdcGuid",
		strings.ToUpper(sdcGUID)); err != nil {
		fields["sdcGuid"] = sdcGUID
		return goof.WithFieldsE(fields, "error finding sdc", err)
	}

	log.WithFields(fields).Info("storage driver initialized")

	return nil
}

func (d *driver) Type(ctx types.Context) (types.StorageType, error) {
	return types.Block, nil
}

// NextDeviceInfo returns the information about the driver's next available
// device workflow.

func (d *driver) NextDeviceInfo(ctx types.Context) (*types.NextDeviceInfo, error) {
	return nil, nil
}

func (d *driver) InstanceInspect(
	ctx types.Context,
	opts types.Store) (*types.Instance, error) {
	iid, _ := d.InstanceID(ctx, opts)
	curInstance := &types.Instance{InstanceID: iid}
	return curInstance, nil
}

func (d *driver) Volumes(ctx types.Context,
	opts *types.VolumesOpts) ([]*types.Volume, error) {

	sdcMappedVolumes, err := goscaleio.GetLocalVolumeMap()
	if err != nil {
		return []*types.Volume{}, err
	}

	mapStoragePoolName, err := d.getStoragePoolIDs()
	if err != nil {
		return []*types.Volume{}, err
	}

	mapProtectionDomainName, err := d.getProtectionDomainIDs()
	if err != nil {
		return []*types.Volume{}, err
	}

	getStoragePoolName := func(ID string) string {
		if pool, ok := mapStoragePoolName[ID]; ok {
			return pool.Name
		}
		return ""
	}

	getProtectionDomainName := func(poolID string) string {
		var ok bool
		var pool *goscaleioTypes.StoragePool

		if pool, ok = mapStoragePoolName[poolID]; !ok {
			return ""
		}

		if protectionDomain, ok := mapProtectionDomainName[pool.ProtectionDomainID]; ok {
			return protectionDomain.Name
		}
		return ""
	}

	sdcDeviceMap := make(map[string]*goscaleio.SdcMappedVolume)
	for _, sdcMappedVolume := range sdcMappedVolumes {
		sdcDeviceMap[sdcMappedVolume.VolumeID] = sdcMappedVolume
	}

	volumes, err := d.getVolume("", "", false)
	if err != nil {
		return []*types.Volume{}, err
	}

	var volumesSD []*types.Volume
	for _, volume := range volumes {
		var attachmentsSD []*types.VolumeAttachment
		for _, attachment := range volume.MappedSdcInfo {
			var deviceName string
			if attachment.SdcID == d.sdc.Sdc.ID {
				if _, exists := sdcDeviceMap[volume.ID]; exists {
					deviceName = sdcDeviceMap[volume.ID].SdcDevice
				}
			}
			instanceID := &types.InstanceID{
				ID: attachment.SdcID,
			}
			attachmentSD := &types.VolumeAttachment{
				VolumeID:   volume.ID,
				InstanceID: instanceID,
				DeviceName: deviceName,
				Status:     "",
			}
			attachmentsSD = append(attachmentsSD, attachmentSD)
		}

		var IOPS int64
		if len(volume.MappedSdcInfo) > 0 {
			IOPS = int64(volume.MappedSdcInfo[0].LimitIops)
		}
		volumeSD := &types.Volume{
			Name:             volume.Name,
			ID:               volume.ID,
			AvailabilityZone: getProtectionDomainName(volume.StoragePoolID),
			Status:           "",
			Type:             getStoragePoolName(volume.StoragePoolID),
			IOPS:             IOPS,
			Size:             int64(volume.SizeInKb / 1024 / 1024),
			Attachments:      attachmentsSD,
		}
		volumesSD = append(volumesSD, volumeSD)
	}

	return volumesSD, nil
}

func (d *driver) VolumeInspect(
	ctx types.Context,
	volumeID string, opts *types.VolumeInspectOpts) (*types.Volume, error) {

	sdcMappedVolumes, err := goscaleio.GetLocalVolumeMap()
	if err != nil {
		return &types.Volume{}, err
	}

	mapStoragePoolName, err := d.getStoragePoolIDs()
	if err != nil {
		return &types.Volume{}, err
	}

	mapProtectionDomainName, err := d.getProtectionDomainIDs()
	if err != nil {
		return &types.Volume{}, err
	}

	getStoragePoolName := func(ID string) string {
		if pool, ok := mapStoragePoolName[ID]; ok {
			return pool.Name
		}
		return ""
	}

	getProtectionDomainName := func(poolID string) string {
		var ok bool
		var pool *goscaleioTypes.StoragePool

		if pool, ok = mapStoragePoolName[poolID]; !ok {
			return ""
		}

		if protectionDomain, ok := mapProtectionDomainName[pool.ProtectionDomainID]; ok {
			return protectionDomain.Name
		}
		return ""
	}

	sdcDeviceMap := make(map[string]*goscaleio.SdcMappedVolume)
	for _, sdcMappedVolume := range sdcMappedVolumes {
		sdcDeviceMap[sdcMappedVolume.VolumeID] = sdcMappedVolume
	}
	volumes, err := d.getVolume(volumeID, "", false)
	if err != nil {
    log.Warn(err)
		return &types.Volume{}, err
	}

	var volumesSD []*types.Volume
	for _, volume := range volumes {
		var attachmentsSD []*types.VolumeAttachment
		for _, attachment := range volume.MappedSdcInfo {
			var deviceName string
			if attachment.SdcID == d.sdc.Sdc.ID {
				if _, exists := sdcDeviceMap[volume.ID]; exists {
					deviceName = sdcDeviceMap[volume.ID].SdcDevice
				}
			}
			instanceID := &types.InstanceID{
				ID: attachment.SdcID,
			}
			attachmentSD := &types.VolumeAttachment{
				VolumeID:   volume.ID,
				InstanceID: instanceID,
				DeviceName: deviceName,
				Status:     "",
			}
			attachmentsSD = append(attachmentsSD, attachmentSD)
		}

		var IOPS int64
		if len(volume.MappedSdcInfo) > 0 {
			IOPS = int64(volume.MappedSdcInfo[0].LimitIops)
		}
		volumeSD := &types.Volume{
			Name:             volume.Name,
			ID:               volume.ID,
			AvailabilityZone: getProtectionDomainName(volume.StoragePoolID),
			Status:           "",
			Type:             getStoragePoolName(volume.StoragePoolID),
			IOPS:             IOPS,
			Size:             int64(volume.SizeInKb / 1024 / 1024),
			Attachments:      attachmentsSD,
		}
		volumesSD = append(volumesSD, volumeSD)
	}
	if len(volumesSD) == 0 {
		return &types.Volume{}, nil
	}
	return volumesSD[0], nil
}

func (d *driver) VolumeCreate(
	ctx types.Context,
	name string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {

	if opts == nil {
		opts = &types.VolumeCreateOpts{}
	}

	if opts.Type == nil {
		var optType string = ""
		opts.Type = &optType
	}
	if opts.IOPS == nil {
		var iops int64 = 0
		opts.IOPS = &iops
	}
	if opts.Size == nil {
		var size int64 = 1
		opts.Size = &size
	}
	if opts.AvailabilityZone == nil {
		var availabilityZone string = ""
		opts.AvailabilityZone = &availabilityZone
	}

	// notUsed bool,volumeName, volumeID, snapshotID, volumeType string,
	// IOPS, size int64, availabilityZone string) (*types.VolumeResp, error)
	if name == "" {
		return &types.Volume{}, goof.WithFields(eff(map[string]interface{}{
			"moduleName": ctx}),
			"no volume name specified")
	}

	volumes, err := d.getVolume("", name, false)
	if err != nil {
		return &types.Volume{}, err
	}

	if len(volumes) > 0 {
		return &types.Volume{}, goof.WithFields(eff(map[string]interface{}{
			"moduleName": ctx,
			"volumeName": name}),
			"volume name already exists")
	}

	resp, err := d.createVolume(ctx,
		false, name, "",
		*opts.Type, *opts.IOPS, *opts.Size, *opts.AvailabilityZone)
	if err != nil {
		return &types.Volume{}, err
	}

	volumes, err = d.getVolume(resp.ID, "", false)
	if err != nil {
		return &types.Volume{}, err
	}

	createdVolume, err := d.VolumeInspect(ctx, resp.ID, nil)
	if err != nil {
		return &types.Volume{}, err
	}

	log.WithFields(log.Fields{
		"moduleName": ctx,
		"provider":   "scaleIO",
		"volume":     createdVolume,
	}).Debug("created volume")
	return createdVolume, nil

}

func (d *driver) VolumeCreateFromSnapshot(
	ctx types.Context,
	snapshotID, volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {
	return nil, nil
}

func (d *driver) VolumeCopy(
	ctx types.Context,
	volumeID, volumeName string,
	opts types.Store) (*types.Volume, error) {
	return nil, nil
}

func (d *driver) VolumeSnapshot(
	ctx types.Context,
	volumeID, snapshotName string,
	opts types.Store) (*types.Snapshot, error) {
	return nil, nil
}

func (d *driver) VolumeRemove(
	ctx types.Context,
	volumeID string,
	opts types.Store) error {

	fields := eff(map[string]interface{}{
		"volumeId": volumeID,
	})

	if volumeID == "" {
		return goof.WithFields(fields, "volumeId is required")
	}

	var err error
	var volumes []*goscaleioTypes.Volume

	if volumes, err = d.getVolume(volumeID, "", false); err != nil {
		return goof.WithFieldsE(fields, "error getting volume", err)
	}

	targetVolume := goscaleio.NewVolume(d.client)
	targetVolume.Volume = volumes[0]

	if err = targetVolume.RemoveVolume("ONLY_ME"); err != nil {
		return goof.WithFieldsE(fields, "error removing volume", err)
	}

	log.WithFields(fields).Debug("removed volume")
	return nil
}

func (d *driver) VolumeAttach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeAttachOpts) (*types.Volume, error) {

	fields := eff(map[string]interface{}{
		"volumeId":   volumeID,
		"instanceId": d.InstanceID,
	})

	if volumeID == "" {
		return nil, goof.WithFields(fields, "volumeId is required")
	}

	mapVolumeSdcParam := &goscaleioTypes.MapVolumeSdcParam{
		SdcID: d.sdc.Sdc.ID,
		AllowMultipleMappings: "false",
		AllSdcs:               "",
	}

	volumes, err := d.getVolume(volumeID, "", false)
	if err != nil {
		return nil, goof.WithFieldsE(fields, "error getting volume", err)
	}

	if len(volumes) == 0 {
		return nil, goof.WithFields(fields, "no volumes returned")
	}

	targetVolume := goscaleio.NewVolume(d.client)
	targetVolume.Volume = volumes[0]

	err = targetVolume.MapVolumeSdc(mapVolumeSdcParam)
	if err != nil {
		return nil, goof.WithFieldsE(fields, "error mapping volume sdc", err)
	}

	_, err = d.waitMount(ctx, volumes[0].ID, opts.Opts)
	if err != nil {
		fields["volumeId"] = volumes[0].ID
		return nil, goof.WithFieldsE(
			fields, "error waiting on volume to mount", err)
	}

	instanceID, _ := d.InstanceID(ctx, opts.Opts)
	volumeInspectOpts := &types.VolumeInspectOpts{true, opts.Opts}
	_, err = d.GetVolumeAttach(ctx, volumeID, instanceID.ID, volumeInspectOpts)
	if err != nil {
		return nil, goof.WithFieldsE(
			fields, "error getting volume attachments", err)
	}

	log.WithFields(log.Fields{
		"provider":   Name,
		"volumeId":   volumeID,
		"instanceId": instanceID.ID,
	}).Debug("attached volume to instance")

	attachedVol, err := d.VolumeInspect(ctx, volumeID, volumeInspectOpts)
	if err != nil {
		return nil, goof.WithFieldsE(fields, "error getting volume", err)
	}

	return attachedVol, nil
}

func (d *driver) VolumeDetach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeDetachOpts) (*types.Volume, error) {
	fields := eff(map[string]interface{}{
		"moduleName": ctx,
		"volumeId":   volumeID,
	})

	if volumeID == "" {
		return &types.Volume{}, goof.WithFields(fields, "volumeId is required")
	}

	volumes, err := d.getVolume(volumeID, "", false)
	if err != nil {
		return &types.Volume{}, goof.WithFieldsE(fields, "error getting volume", err)
	}

	if len(volumes) == 0 {
		return &types.Volume{}, goof.WithFields(fields, "no volumes returned")
	}

	targetVolume := goscaleio.NewVolume(d.client)
	targetVolume.Volume = volumes[0]

	unmapVolumeSdcParam := &goscaleioTypes.UnmapVolumeSdcParam{
		SdcID:                "",
		IgnoreScsiInitiators: "true",
		AllSdcs:              "",
	}

	unmapVolumeSdcParam.SdcID = d.sdc.Sdc.ID

	_ = targetVolume.UnmapVolumeSdc(unmapVolumeSdcParam)

	log.WithFields(log.Fields{
		"moduleName": ctx,
		"provider":   Name,
		"volumeId":   volumeID}).Debug("detached volume")

	return &types.Volume{
    Name: targetVolume.Volume.Name,
    Size: int64(targetVolume.Volume.SizeInKb/1048576),
    Type: targetVolume.Volume.VolumeType,
    ID: targetVolume.Volume.ID,
    Attachments: []*types.VolumeAttachment{},
  }, nil
}

func (d *driver) Snapshots(
	ctx types.Context,
	opts types.Store) ([]*types.Snapshot, error) {
	return nil, nil
}

func (d *driver) SnapshotInspect(
	ctx types.Context,
	snapshotID string,
	opts types.Store) (*types.Snapshot, error) {
	return nil, nil
}

func (d *driver) SnapshotCopy(
	ctx types.Context,
	snapshotID, snapshotName, destinationID string,
	opts types.Store) (*types.Snapshot, error) {
	return nil, nil
}

func (d *driver) SnapshotRemove(
	ctx types.Context,
	snapshotID string,
	opts types.Store) error {
	return nil
}

///////////////////////////////////////////////////////////////////////
////// HELPER FUNCTIONS FOR SCALEIO DRIVER FROM THIS POINT ON /////////
///////////////////////////////////////////////////////////////////////

func shrink(n string) string {
	if len(n) > cc {
		return n[:cc]
	}
	return n
}

func (d *driver) getStoragePoolIDs() (map[string]*goscaleioTypes.StoragePool, error) {
	storagePools, err := d.client.GetStoragePool("")
	if err != nil {
		return nil, err
	}

	mapPoolID := make(map[string]*goscaleioTypes.StoragePool)

	for _, pool := range storagePools {
		mapPoolID[pool.ID] = pool
	}
	return mapPoolID, nil
}

func (d *driver) getProtectionDomainIDs() (map[string]*goscaleioTypes.ProtectionDomain, error) {
	protectionDomains, err := d.system.GetProtectionDomain("")
	if err != nil {
		return nil, err
	}

	mapProtectionDomainID := make(map[string]*goscaleioTypes.ProtectionDomain)

	for _, protectionDomain := range protectionDomains {
		mapProtectionDomainID[protectionDomain.ID] = protectionDomain
	}
	return mapProtectionDomainID, nil
}

func (d *driver) getVolume(
	volumeID, volumeName string, getSnapshots bool) ([]*goscaleioTypes.Volume, error) {

	volumeName = shrink(volumeName)

	volumes, err := d.client.GetVolume("", volumeID, "", volumeName, getSnapshots)
	if err != nil {
		return []*goscaleioTypes.Volume{}, err
	}
	return volumes, nil
}

func (d *driver) createVolume(ctx types.Context,
	notUsed bool,
	volumeName, volumeID, volumeType string,
	IOPS, size int64, availabilityZone string) (*goscaleioTypes.VolumeResp, error) {

	volumeName = shrink(volumeName)

	fields := eff(map[string]interface{}{
		"moduleName":       ctx,
		"volumeID":         volumeID,
		"volumeName":       volumeName,
		"volumeType":       volumeType,
		"IOPS":             IOPS,
		"size":             size,
		"availabilityZone": availabilityZone,
	})

	volumeParam := &goscaleioTypes.VolumeParam{
		Name:           volumeName,
		VolumeSizeInKb: strconv.Itoa(int(size) * 1024 * 1024),
		VolumeType:     d.thinOrThick(),
	}

	if volumeType == "" {
		volumeType = d.storagePool.StoragePool.Name
		fields["volumeType"] = volumeType
	}

	volumeResp, err := d.client.CreateVolume(volumeParam, volumeType)
	if err != nil {
		return &goscaleioTypes.VolumeResp{}, goof.WithFieldsE(fields, "error creating volume", err)
	}

	return volumeResp, nil
}

//TODO change provider to be dynamic...

func eff(fields goof.Fields) map[string]interface{} {
	errFields := map[string]interface{}{
		"provider": "scaleio",
	}
	if fields != nil {
		for k, v := range fields {
			errFields[k] = v
		}
	}
	return errFields
}

func (d *driver) waitMount(ctx types.Context, volumeID string, opts types.Store) (*goscaleio.SdcMappedVolume, error) {

	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(10 * time.Second)
		timeout <- true
	}()

	successCh := make(chan *goscaleio.SdcMappedVolume, 1)
	errorCh := make(chan error, 1)
	go func(volumeID string) {
		log.WithField("provider", Name).Debug("waiting for volume mount")
		for {
			sdcMappedVolumes, err := goscaleio.GetLocalVolumeMap()
			if err != nil {
				errorCh <- goof.WithFieldE(
					"provider", Name,
					"problem getting local volume mappings", err)
				return
			}

			sdcMappedVolume := &goscaleio.SdcMappedVolume{}
			var foundVolume bool
			for _, sdcMappedVolume = range sdcMappedVolumes {
				if sdcMappedVolume.VolumeID ==
					volumeID && sdcMappedVolume.SdcDevice != "" {
					foundVolume = true
					break
				}
			}

			if foundVolume {
				successCh <- sdcMappedVolume
				return
			}
			time.Sleep(100 * time.Millisecond)
		}

	}(volumeID)

	select {
	case sdcMappedVolume := <-successCh:
		log.WithFields(log.Fields{
			"moduleName": ctx,
			"provider":   Name,
			"volumeId":   sdcMappedVolume.VolumeID,
			"volume":     sdcMappedVolume.SdcDevice,
		}).Debug("got sdcMappedVolume")
		return sdcMappedVolume, nil
	case err := <-errorCh:
		return &goscaleio.SdcMappedVolume{}, err
	case <-timeout:
		return &goscaleio.SdcMappedVolume{}, goof.WithFields(
			ef(), "timed out waiting for mount")
	}

}

func (d *driver) GetVolumeAttach(
	ctx types.Context, volumeID, instanceID string, opts *types.VolumeInspectOpts) ([]*types.VolumeAttachment, error) {

	fields := eff(map[string]interface{}{
		"volumeId":   volumeID,
		"instanceId": instanceID,
	})

	if volumeID == "" {
		return []*types.VolumeAttachment{},
			goof.WithFields(fields, "volumeId is required")
	}
	volume, err := d.VolumeInspect(ctx, volumeID, opts)
	if err != nil {
		return []*types.VolumeAttachment{},
			goof.WithFieldsE(fields, "error getting volume", err)
	}

	if instanceID != "" {
		var attached bool
		for _, volumeAttachment := range volume.Attachments {
			if volumeAttachment.InstanceID.ID == instanceID {
				return volume.Attachments, nil
			}
		}
		if !attached {
			return []*types.VolumeAttachment{}, nil
		}
	}
	return volume.Attachments, nil
}

///////////////////////////////////////////////////////////////////////
//////                  CONFIG HELPER STUFF                   /////////
///////////////////////////////////////////////////////////////////////

func (d *driver) endpoint() string {
	return d.StorageExecutor.Config.GetString("scaleio.endpoint")
}

func (d *driver) insecure() bool {
	return d.StorageExecutor.Config.GetBool("scaleio.insecure")
}

func (d *driver) useCerts() bool {
	return d.StorageExecutor.Config.GetBool("scaleio.useCerts")
}

func (d *driver) userID() string {
	return d.StorageExecutor.Config.GetString("scaleio.userID")
}

func (d *driver) userName() string {
	return d.StorageExecutor.Config.GetString("scaleio.userName")
}

func (d *driver) password() string {
	return d.StorageExecutor.Config.GetString("scaleio.password")
}

func (d *driver) systemID() string {
	return d.StorageExecutor.Config.GetString("scaleio.systemID")
}

func (d *driver) systemName() string {
	return d.StorageExecutor.Config.GetString("scaleio.systemName")
}

func (d *driver) protectionDomainID() string {
	return d.StorageExecutor.Config.GetString("scaleio.protectionDomainID")
}

func (d *driver) protectionDomainName() string {
	return d.StorageExecutor.Config.GetString("scaleio.protectionDomainName")
}

func (d *driver) storagePoolID() string {
	return d.StorageExecutor.Config.GetString("scaleio.storagePoolID")
}

func (d *driver) storagePoolName() string {
	return d.StorageExecutor.Config.GetString("scaleio.storagePoolName")
}

func (d *driver) thinOrThick() string {
	thinOrThick := d.StorageExecutor.Config.GetString("scaleio.thinOrThick")
	if thinOrThick == "" {
		return "ThinProvisioned"
	}
	return thinOrThick
}

func (d *driver) version() string {
	return d.StorageExecutor.Config.GetString("scaleio.version")
}

func ef() goof.Fields {
	return goof.Fields{
		"provider": Name,
	}
}
