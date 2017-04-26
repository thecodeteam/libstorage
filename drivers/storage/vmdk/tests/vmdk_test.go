// +build linux darwin
// +build !libstorage_storage_driver libstorage_storage_driver_vmdk

package tests

import (
	"io/ioutil"
	"os"
	"strconv"
	"testing"

	log "github.com/Sirupsen/logrus"

	"github.com/vmware/docker-volume-vsphere/vmdk_plugin/drivers/vmdk/vmdkops"

	"github.com/codedellemc/libstorage/api/context"
	"github.com/codedellemc/libstorage/api/registry"
	"github.com/codedellemc/libstorage/api/server"
	"github.com/codedellemc/libstorage/api/types"
	"github.com/codedellemc/libstorage/api/utils"

	// load the driver
	_ "github.com/codedellemc/libstorage/drivers/storage/vmdk/storage"
)

var tCtx types.Context
var vName string
var d types.StorageDriver

func skipTests() bool {
	travis, _ := strconv.ParseBool(os.Getenv("TRAVIS"))
	noTest, _ := strconv.ParseBool(os.Getenv("TEST_SKIP_VMDK"))
	return travis || noTest
}

func TestMain(m *testing.M) {
	if p, err := strconv.Atoi(os.Getenv("VMDK_PORT")); err == nil {
		vmdkops.EsxPort = p
	} else {
		vmdkops.EsxPort = 1019
	}

	if vName = os.Getenv("VMDK_NAME"); vName == "" {
		vName = "vmdkops"
	}

	log.SetLevel(log.DebugLevel)

	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)

	tCtx = context.Background()
	pathConfig := utils.NewPathConfig(tCtx, tmpDir, "")
	tCtx = context.WithValue(tCtx, context.PathConfigKey, pathConfig)
	registry.ProcessRegisteredConfigs(tCtx)

	server.CloseOnAbort()

	d, _ = registry.NewStorageDriver("vmdk")
	d.Init(tCtx, nil)

	os.Exit(m.Run())
}

func TestVolumeInspect(t *testing.T) {
	if skipTests() {
		t.SkipNow()
	}

	v, err := d.VolumeInspect(
		tCtx, vName, &types.VolumeInspectOpts{Opts: utils.NewStore()})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", v)
}

func TestVolumeCreate(t *testing.T) {
	if skipTests() {
		t.SkipNow()
	}

	v, err := d.VolumeCreate(
		tCtx, vName, &types.VolumeCreateOpts{Opts: utils.NewStore()})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", v)
}

func TestVolumeAttach(t *testing.T) {
	if skipTests() {
		t.SkipNow()
	}

	v, tok, err := d.VolumeAttach(
		tCtx, vName, &types.VolumeAttachOpts{Opts: utils.NewStore()})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("tok=%v, %+v", tok, v)
}

func TestVolumeDetach(t *testing.T) {
	if skipTests() {
		t.SkipNow()
	}

	v, err := d.VolumeDetach(
		tCtx, vName, &types.VolumeDetachOpts{Opts: utils.NewStore()})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", v)
}

func TestVolumeRemove(t *testing.T) {
	if skipTests() {
		t.SkipNow()
	}

	err := d.VolumeRemove(
		tCtx, vName, &types.VolumeRemoveOpts{Opts: utils.NewStore()})
	if err != nil {
		t.Fatal(err)
	}
}

func TestVolumeList(t *testing.T) {
	if skipTests() {
		t.SkipNow()
	}

	vols, err := d.Volumes(tCtx, &types.VolumesOpts{Opts: utils.NewStore()})
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range vols {
		t.Logf("%+v", v)
	}
}
