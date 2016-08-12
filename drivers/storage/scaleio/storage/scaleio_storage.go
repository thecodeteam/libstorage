package storage

import (
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/goof"
	sio "github.com/emccode/goscaleio"
	siotypes "github.com/emccode/goscaleio/types/v1"

	"github.com/emccode/libstorage/api/context"
	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/drivers/storage/scaleio"
)

const (
	cc = 31
)

type driver struct {
	config           gofig.Config
	client           *sio.Client
	system           *siotypes.System
	protectionDomain *siotypes.ProtectionDomain
	storagePool      *siotypes.StoragePool
}

func init() {
	registry.RegisterStorageDriver(scaleio.Name, newDriver)
}

func newDriver() types.StorageDriver {
	return &driver{}
}

func (d *driver) Name() string {
	return scaleio.Name
}

func (d *driver) Init(context types.Context, config gofig.Config) error {
	d.config = config
	fields := eff(map[string]interface{}{
		"endpoint": d.endpoint(),
		"insecure": d.insecure(),
		"useCerts": d.useCerts(),
	})

	log.WithFields(fields).Debug("starting scaleio driver")

	var err error

	if d.client, err = sio.NewClientWithArgs(
		d.endpoint(),
		d.version(),
		d.insecure(),
		d.useCerts()); err != nil {
		return goof.WithFieldsE(fields, "error constructing new client", err)
	}

	if err = d.auth(); err != nil {
		fields["userName"] = d.userName()
		if d.password() != "" {
			fields["password"] = "******"
		}
		log.WithFields(fields).Debug(err.Error())
		return goof.WithFieldsE(fields, "error authenticating", err)
	}

	if d.system, err = d.findSystem(d.systemID(), d.systemName()); err != nil {
		fields["systemId"] = d.systemID()
		fields["systemName"] = d.systemName()
		log.WithFields(fields).Debug(err.Error())
		return goof.WithFieldsE(fields, "error finding system", err)
	}

	if d.protectionDomain, err = d.findProtectionDomain(
		d.protectionDomainID(),
		d.protectionDomainName()); err != nil {

		fields["domainId"] = d.protectionDomainID()
		fields["domainName"] = d.protectionDomainName()
		log.WithFields(fields).Debug(err.Error())
		return goof.WithFieldsE(fields,
			"error finding protection domain", err)
	}

	if d.storagePool, err = d.findStoragePool(
		d.storagePoolID(),
		d.storagePoolName()); err != nil {

		fields["storagePoolId"] = d.storagePoolID()
		fields["storagePoolName"] = d.storagePoolName()
		log.WithFields(fields).Debug(err.Error())
		return goof.WithFieldsE(fields, "error finding storage pool", err)
	}

	log.WithFields(fields).Info("storage driver initialized")

	return nil
}

func (d *driver) Type(ctx types.Context) (types.StorageType, error) {
	return types.Block, nil
}

func (d *driver) NextDeviceInfo(
	ctx types.Context) (*types.NextDeviceInfo, error) {
	return nil, nil
}

func (d *driver) InstanceInspect(
	ctx types.Context,
	opts types.Store) (*types.Instance, error) {

	iid := context.MustInstanceID(ctx)
	if iid.ID != "" {
		return &types.Instance{InstanceID: iid}, nil
	}

	var (
		err     error
		sdcGUID string
		sdc     *siotypes.Sdc
	)

	if err = iid.UnmarshalMetadata(&sdcGUID); err != nil {
		return nil, err
	}

	sdcGUID = strings.ToUpper(sdcGUID)
	if sdc, err = d.getSdcByGUID(sdcGUID); err != nil {
		return nil, scaleio.ErrFindingSDC(sdcGUID, err)
	}

	if sdc != nil {
		return &types.Instance{
			InstanceID: &types.InstanceID{
				ID:     sdc.ID,
				Driver: d.Name(),
			},
		}, nil
	}

	return nil, scaleio.ErrNoSDCGUID
}

func (d *driver) Volumes(
	ctx types.Context,
	opts *types.VolumesOpts) ([]*types.Volume, error) {

	sdcMappedVolumes := make(map[string]string)
	if opts.Attachments {
		if ld, ok := context.LocalDevices(ctx); ok {
			sdcMappedVolumes = ld.DeviceMap
		}
	}

	volumes, err := d.getVolumes()
	if err != nil {
		return nil, err
	}

	var volumesSD []*types.Volume
	for _, volume := range volumes {
		var attachmentsSD []*types.VolumeAttachment
		if opts != nil && opts.Attachments {
			for _, attachment := range volume.MappedSdcInfo {
				var deviceName string
				if _, exists := sdcMappedVolumes[volume.ID]; exists {
					deviceName = sdcMappedVolumes[volume.ID]
				}
				instanceID := &types.InstanceID{
					ID:     attachment.SdcID,
					Driver: d.Name(),
				}
				attachmentSD := &types.VolumeAttachment{
					VolumeID:   volume.ID,
					InstanceID: instanceID,
					DeviceName: deviceName,
					Status:     "",
				}
				attachmentsSD = append(attachmentsSD, attachmentSD)
			}
		}

		var IOPS int64
		if len(volume.MappedSdcInfo) > 0 {
			IOPS = int64(volume.MappedSdcInfo[0].LimitIops)
		}
		volumeSD := &types.Volume{
			Name:             volume.Name,
			ID:               volume.ID,
			AvailabilityZone: d.protectionDomain.Name,
			Status:           "",
			Type:             d.storagePool.Name,
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
	volumeID string,
	opts *types.VolumeInspectOpts) (*types.Volume, error) {

	if volumeID == "" {
		return nil, goof.New("no volumeID specified")
	}

	sdcMappedVolumes := make(map[string]string)
	if opts.Attachments {
		if ld, ok := context.LocalDevices(ctx); ok {
			sdcMappedVolumes = ld.DeviceMap
		}
	}

	volume, err := d.getVolumeByID(volumeID)
	if err != nil {
		return nil, err
	}

	var attachmentsSD []*types.VolumeAttachment
	if opts != nil && opts.Attachments {
		for _, attachment := range volume.MappedSdcInfo {
			var deviceName string
			if _, exists := sdcMappedVolumes[volume.ID]; exists {
				deviceName = sdcMappedVolumes[volume.ID]
			}
			instanceID := &types.InstanceID{
				ID:     attachment.SdcID,
				Driver: d.Name(),
			}
			attachmentSD := &types.VolumeAttachment{
				VolumeID:   volume.ID,
				InstanceID: instanceID,
				DeviceName: deviceName,
				Status:     "",
			}
			attachmentsSD = append(attachmentsSD, attachmentSD)
		}
	}

	var IOPS int64
	if len(volume.MappedSdcInfo) > 0 {
		IOPS = int64(volume.MappedSdcInfo[0].LimitIops)
	}
	volumeSD := &types.Volume{
		Name:             volume.Name,
		ID:               volume.ID,
		AvailabilityZone: d.protectionDomain.Name,
		Status:           "",
		Type:             d.storagePool.Name,
		IOPS:             IOPS,
		Size:             int64(volume.SizeInKb / 1024 / 1024),
		Attachments:      attachmentsSD,
	}

	return volumeSD, nil
}

func (d *driver) VolumeCreate(ctx types.Context, volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {

	fields := eff(map[string]interface{}{
		"volumeName": volumeName,
		"opts":       opts,
	})

	log.WithFields(fields).Debug("creating volume")

	volume := &types.Volume{}

	if opts.AvailabilityZone != nil {
		volume.AvailabilityZone = *opts.AvailabilityZone
	}
	if opts.Type != nil {
		volume.Type = *opts.Type
	}
	if opts.Size != nil {
		volume.Size = *opts.Size
	}
	if opts.IOPS != nil {
		volume.IOPS = *opts.IOPS
	}

	volID, err := d.createVolume(ctx, volumeName, volume)
	if err != nil {
		return nil, err
	}

	return d.VolumeInspect(ctx, volID, &types.VolumeInspectOpts{
		Attachments: true,
	})
}

func (d *driver) VolumeCreateFromSnapshot(
	ctx types.Context,
	snapshotID, volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {

	// notUsed bool,volumeName, volumeID, snapshotID, volumeType string,
	// IOPS, size int64, availabilityZone string) (*types.VolumeResp, error)
	if volumeName == "" {
		return nil, goof.New("no volume name specified")
	}

	volumes, err := d.findVolumes("", volumeName, false)
	if err != nil {
		return nil, err
	}

	if len(volumes) > 0 {
		return nil, goof.WithFields(eff(map[string]interface{}{
			"volumeName": volumeName}),
			"volume name already exists")
	}

	return d.VolumeCreate(ctx, volumeName, opts)
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

	err := d.client.RemoveVolume(volumeID, sio.RemoveMode.OnlyMe)
	if d.isReauthNeeded(err) {
		err = d.client.RemoveVolume(volumeID, sio.RemoveMode.OnlyMe)
	}
	if err != nil {
		return goof.WithFieldsE(fields, "error removing volume", err)
	}

	log.WithFields(fields).Debug("removed volume")
	return nil
}

func (d *driver) VolumeAttach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeAttachOpts) (*types.Volume, string, error) {

	iid := context.MustInstanceID(ctx)

	mapVolumeSdcParam := &siotypes.MapVolumeSdcParam{
		SdcID: iid.ID,
		AllowMultipleMappings: "false",
		AllSdcs:               "",
	}

	if !opts.Force {
		if err := d.addMappedSdc(volumeID, mapVolumeSdcParam); err != nil {
			return nil, "", goof.WithError("error mapping volume sdc", err)
		}
	} else {
		vol, err := d.VolumeInspect(
			ctx, volumeID, &types.VolumeInspectOpts{
				Attachments: true,
			})
		if err != nil {
			return nil, "", goof.WithError("error getting volume", err)
		}
		if len(vol.Attachments) > 0 && opts.Force {
			param := &siotypes.UnmapVolumeSdcParam{
				IgnoreScsiInitiators: "true",
				AllSdcs:              "true",
			}
			if err := d.removeMappedSdc(volumeID, param); err != nil {
				return nil, "", goof.WithError(
					"error removing sdc mapping", err)
			}
		}
		if err := d.addMappedSdc(volumeID, mapVolumeSdcParam); err != nil {
			return nil, "", goof.WithError("error mapping volume sdc", err)
		}
	}

	attachedVol, err := d.VolumeInspect(
		ctx, volumeID, &types.VolumeInspectOpts{
			Attachments: true,
			Opts:        opts.Opts,
		})
	if err != nil {
		return nil, "", goof.WithError("error getting volume", err)
	}

	return attachedVol, attachedVol.ID, nil
}

func (d *driver) VolumeDetach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeDetachOpts) (*types.Volume, error) {

	iid := context.MustInstanceID(ctx)

	unmapVolumeSdcParam := &siotypes.UnmapVolumeSdcParam{
		SdcID:                "",
		IgnoreScsiInitiators: "true",
		AllSdcs:              "",
	}

	if opts.Force {
		unmapVolumeSdcParam.AllSdcs = "true"
	} else {
		unmapVolumeSdcParam.SdcID = iid.ID
	}

	err := d.client.RemoveMappedSdc(volumeID, unmapVolumeSdcParam)
	if d.isReauthNeeded(err) {
		err = d.client.RemoveMappedSdc(volumeID, unmapVolumeSdcParam)
	}
	if err != nil {
		return nil, err
	}

	vol, err := d.VolumeInspect(ctx, volumeID, &types.VolumeInspectOpts{
		Attachments: true,
	})
	if err != nil {
		return nil, err
	}

	return vol, nil
}

func (d *driver) VolumeDetachAll(
	ctx types.Context,
	volumeID string,
	opts types.Store) error {
	return nil
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

// isReauthNeeded tests the passed error for unauthorized, if so, reauth.
func (d *driver) isReauthNeeded(err error) bool {
	if err == nil {
		log.WithField("reauth", "false").
			Debug("reauth assertion skipped")
		return false
	}

	// test for unauthorized
	if err, ok := err.(*siotypes.Error); ok {
		if err.HTTPStatusCode == 401 ||
			strings.ToLower(err.Message) == "unauthorized" {
			log.WithField("reauth", "true").Debug("reauthorizing API access")
			if authErr := d.auth(); authErr != nil {
				log.WithField("reauth", "true").
					WithError(err).Error("reauthorizing failed")
				return false
			}
			return true
		}
	}
	log.WithField("reauth", "false").Debugf("not auth error: type %T", err)
	return false
}

func (d *driver) auth() error {
	return d.client.Authenticate(
		&sio.ConfigConnect{
			Endpoint: d.endpoint(),
			Version:  d.version(),
			Username: d.userName(),
			Password: d.password()})
}

func (d *driver) getSdcByGUID(guid string) (*siotypes.Sdc, error) {
	sdc, err := d.client.GetSdcByGUID(guid)
	if d.isReauthNeeded(err) {
		sdc, err = d.client.GetSdcByGUID(guid)
	}
	return sdc, err
}

func (d *driver) getVolumes() ([]*siotypes.Volume, error) {
	volumes, err := d.client.GetVolumes()
	if d.isReauthNeeded(err) {
		volumes, err = d.client.GetVolumes()
	}
	return volumes, err
}

func (d *driver) getVolumeByID(id string) (*siotypes.Volume, error) {
	vol, err := d.client.GetVolumeByID(id)
	if d.isReauthNeeded(err) {
		vol, err = d.client.GetVolumeByID(id)
	}
	return vol, err
}

func (d *driver) addMappedSdc(volID string,
	param *siotypes.MapVolumeSdcParam) error {
	err := d.client.AddMappedSdc(volID, param)
	if d.isReauthNeeded(err) {
		err = d.client.AddMappedSdc(volID, param)
	}
	return err
}

func (d *driver) removeMappedSdc(volID string,
	param *siotypes.UnmapVolumeSdcParam) error {
	err := d.client.RemoveMappedSdc(volID, param)
	if d.isReauthNeeded(err) {
		err = d.client.RemoveMappedSdc(volID, param)
	}
	return err
}

func (d *driver) findSystem(id, name string) (*siotypes.System, error) {
	if id != "" {
		sys, err := d.client.GetSystemByID(id)
		if d.isReauthNeeded(err) {
			sys, err = d.client.GetSystemByID(id)
		}
		return sys, err
	} else if name != "" {
		sys, err := d.client.GetSystemByName(name)
		if d.isReauthNeeded(err) {
			sys, err = d.client.GetSystemByName(name)
		}
		return sys, err
	}
	return nil, goof.New("missing id or name")
}

func (d *driver) findStoragePool(
	id, name string) (*siotypes.StoragePool, error) {
	if id != "" {
		pool, err := d.client.GetStoragePoolByID(id)
		if d.isReauthNeeded(err) {
			pool, err = d.client.GetStoragePoolByID(id)
		}
		return pool, err
	} else if name != "" {
		pool, err := d.client.GetStoragePoolByName(name)
		if d.isReauthNeeded(err) {
			pool, err = d.client.GetStoragePoolByName(name)
		}
		return pool, err
	}
	return nil, goof.New("missing id or name")
}

func (d *driver) findProtectionDomain(
	id, name string) (*siotypes.ProtectionDomain, error) {
	if id != "" {
		domain, err := d.client.GetProtectionDomainByID(id)
		if d.isReauthNeeded(err) {
			domain, err = d.client.GetProtectionDomainByID(id)
		}
		return domain, err
	} else if name != "" {
		domain, err := d.client.GetProtectionDomainByName(name)
		if d.isReauthNeeded(err) {
			domain, err = d.client.GetProtectionDomainByName(name)
		}
		return domain, err
	}
	return nil, goof.New("missing id or name")
}

func (d *driver) getStoragePoolIDs() (
	map[string]*siotypes.StoragePool, error) {
	storagePools, err := d.client.GetStoragePools()
	if d.isReauthNeeded(err) {
		storagePools, err = d.client.GetStoragePools()
	}
	if err != nil {
		return nil, err
	}

	mapPoolID := make(map[string]*siotypes.StoragePool)

	for _, pool := range storagePools {
		mapPoolID[pool.ID] = pool
	}
	return mapPoolID, nil
}

func (d *driver) getProtectionDomainIDs() (
	map[string]*siotypes.ProtectionDomain, error) {
	protectionDomains, err := d.client.GetProtectionDomains()
	if d.isReauthNeeded(err) {
		protectionDomains, err = d.client.GetProtectionDomains()
	}
	if err != nil {
		return nil, err
	}

	mapProtectionDomainID := make(map[string]*siotypes.ProtectionDomain)

	for _, protectionDomain := range protectionDomains {
		mapProtectionDomainID[protectionDomain.ID] = protectionDomain
	}
	return mapProtectionDomainID, nil
}

func (d *driver) findVolumes(
	volumeID, volumeName string, getSnapshots bool) (
	[]*siotypes.Volume, error) {
	volumeName = shrink(volumeName)

	if volumeID != "" {
		vol, err := d.client.GetVolumeByID(volumeID)
		if d.isReauthNeeded(err) {
			vol, err = d.client.GetVolumeByID(volumeID)
		}
		if err != nil {
			return nil, err
		}
		// return snapshot only
		if getSnapshots && vol.AncestorVolumeID == "" {
			return nil, nil
		}
		return []*siotypes.Volume{vol}, nil
	}

	var filtered []*siotypes.Volume
	volumes, err := d.getVolumes()
	if err != nil {
		return nil, err
	}
	for _, vol := range volumes {
		if (vol.Name == volumeName) &&
			(getSnapshots == (vol.AncestorVolumeID != "")) {
			filtered = append(filtered, vol)
		}
	}
	return filtered, nil
}

func (d *driver) createVolume(ctx types.Context, volumeName string,
	vol *types.Volume) (string, error) {

	volumeName = shrink(volumeName)

	fields := eff(map[string]interface{}{
		// "volumeID":         volumeID,
		"volumeName":       volumeName,
		"volumeType":       vol.Type,
		"IOPS":             vol.IOPS,
		"size":             vol.Size,
		"availabilityZone": vol.AvailabilityZone,
	})

	volumeParam := &siotypes.VolumeParam{
		StoragePoolID:  d.storagePool.ID,
		Name:           volumeName,
		VolumeSizeInKb: strconv.Itoa(int(vol.Size) * 1024 * 1024),
		VolumeType:     d.thinOrThick(),
	}

	if vol.Type == "" {
		vol.Type = d.storagePool.Name
		fields["volumeType"] = vol.Type
	}

	volID, err := d.client.CreateVolume(volumeParam)
	if d.isReauthNeeded(err) {
		volID, err = d.client.CreateVolume(volumeParam)
	}
	if err != nil {
		return "", goof.WithFieldsE(fields, "error creating volume", err)
	}

	return volID, nil
}

//TODO change provider to be dynamic...

func eff(fields goof.Fields) map[string]interface{} {
	errFields := map[string]interface{}{
		"provider": "scaleIO",
	}
	if fields != nil {
		for k, v := range fields {
			errFields[k] = v
		}
	}
	return errFields
}

///////////////////////////////////////////////////////////////////////
//////                  CONFIG HELPER STUFF                   /////////
///////////////////////////////////////////////////////////////////////

func (d *driver) endpoint() string {
	return d.config.GetString("scaleio.endpoint")
}

func (d *driver) insecure() bool {
	return d.config.GetBool("scaleio.insecure")
}

func (d *driver) useCerts() bool {
	return d.config.GetBool("scaleio.useCerts")
}

func (d *driver) userID() string {
	return d.config.GetString("scaleio.userID")
}

func (d *driver) userName() string {
	return d.config.GetString("scaleio.userName")
}

func (d *driver) password() string {
	return d.config.GetString("scaleio.password")
}

func (d *driver) systemID() string {
	return d.config.GetString("scaleio.systemID")
}

func (d *driver) systemName() string {
	return d.config.GetString("scaleio.systemName")
}

func (d *driver) protectionDomainID() string {
	return d.config.GetString("scaleio.protectionDomainID")
}

func (d *driver) protectionDomainName() string {
	return d.config.GetString("scaleio.protectionDomainName")
}

func (d *driver) storagePoolID() string {
	return d.config.GetString("scaleio.storagePoolID")
}

func (d *driver) storagePoolName() string {
	return d.config.GetString("scaleio.storagePoolName")
}

func (d *driver) thinOrThick() string {
	thinOrThick := d.config.GetString("scaleio.thinOrThick")
	if thinOrThick == "" {
		return "ThinProvisioned"
	}
	return thinOrThick
}

func (d *driver) version() string {
	return d.config.GetString("scaleio.version")
}
