package scaleio

import (
  "os"
  "strconv"
  "testing"
  "os/exec"
  "strings"

  "github.com/akutz/gofig"
  "github.com/stretchr/testify/assert"

  "github.com/emccode/libstorage/api/server/executors"
  apitests "github.com/emccode/libstorage/api/tests"
  "github.com/emccode/libstorage/api/types"
  "github.com/emccode/libstorage/client"


  // load the  driver
  "github.com/emccode/libstorage/drivers/storage/scaleio"
  "github.com/emccode/libstorage/api/server"
)

var (
  lsxbin string

	lsxLinuxInfo, _   = executors.ExecutorInfoInspect("lsx-linux", false)
	lsxDarwinInfo, _  = executors.ExecutorInfoInspect("lsx-darwin", false)
	lsxWindowsInfo, _ = executors.ExecutorInfoInspect("lsx-windows.exe", false)

	configYAML = []byte(`
libstorage:
  host: tcp://127.0.0.1
  driver: scaleio
  server:
    services:
      scaleio:
        endpoint:             https://192.168.50.12/api
        insecure:             true
        useCerts:             false
        userName:             admin
        password:             Scaleio123
        systemID:             6cfe25856a90658d
        systemName:           cluster1
        protectionDomainID:   6d13747300000000
        protectionDomainName: pdomain
        storagePoolID:        672d836d00000000
        storagePoolName:      pool1
        thinOrThick:          ThinProvisioned
        version:              2.0

scaleio:
  endpoint:             https://192.168.50.12/api
  insecure:             true
  useCerts:             false
  userName:             admin
  password:             Scaleio123
  systemID:             6cfe25856a90658d
  systemName:           cluster1
  protectionDomainID:   6d13747300000000
  protectionDomainName: pdomain
  storagePoolID:        672d836d00000000
  storagePoolName:      pool1
  thinOrThick:          ThinProvisioned
  version:              2.0
`)
)

func init() {
  if travis, _ := strconv.ParseBool(os.Getenv("TRAVIS")); !travis {
    // semaphore.Unlink(types.LSX)
  }
}

func TestMain(m *testing.M) {
  server.CloseOnAbort()
  ec := m.Run()
  os.Exit(ec)
}

func TestClient(t *testing.T) {
  apitests.Run(t, scaleio.Name, configYAML,
    func(config gofig.Config, client client.Client, t *testing.T) {
      iid, err := client.API().InstanceID(nil, scaleio.Name)
      assert.NoError(t, err)
      assert.NotNil(t, iid)
    })
}

func TestInstanceID(t *testing.T) {
  apitests.RunGroup(
    t, scaleio.Name, configYAML,
    (&apitests.InstanceIDTest{
     Driver:   scaleio.Name,
     Expected: getSdcLocalGUID(),
    }).Test)
}

func TestRoot(t *testing.T) {
  apitests.Run(t, scaleio.Name, configYAML, apitests.TestRoot)
}

func TestServices(t *testing.T) {
  tf := func(config gofig.Config, client client.Client, t *testing.T) {
    reply, err := client.API().Services(nil)
    assert.NoError(t, err)
    assert.Equal(t, len(reply), 1)

    _, ok := reply[scaleio.Name]
    assert.True(t, ok)
  }
  apitests.Run(t, scaleio.Name, configYAML, tf)
}

func TestServiceInspect(t *testing.T) {
  tf := func(config gofig.Config, client client.Client, t *testing.T) {
    reply, err := client.API().ServiceInspect(nil, "scaleio")
    assert.NoError(t, err)
    assert.Equal(t, "scaleio", reply.Name)
    assert.Equal(t, "scaleio", reply.Driver.Name)
  }
  apitests.Run(t, scaleio.Name, configYAML, tf)
}

func TestVolumeWorkflow(t *testing.T) {
  volumeEndpointTest := func(config gofig.Config, client client.Client, t *testing.T) {
    serviceVolumeMap, err := client.API().Volumes(nil, false)
    assert.Equal(t, err, nil)
    svMap := serviceVolumeMap["scaleio"]
    for _, v := range svMap {
      assert.NotEqual(t, v.Name, "libstorageTest")
    }
  }
  var volumeID string

  createVolumeTest := func(config gofig.Config, client client.Client, t *testing.T) {

    volumeName := "libstorageTest"
    size := int64(8)

    volumeCreateRequest := &types.VolumeCreateRequest{
      Name:             volumeName,
      Size:             &size,
    }

    created, err := client.API().VolumeCreate(nil, "scaleio",volumeCreateRequest)
    assert.Nil(t, err)
    assert.Equal(t, created.Name, volumeCreateRequest.Name)
    assert.Equal(t, created.Size, *volumeCreateRequest.Size)
    assert.NotNil(t, created.ID)
    volumeID = created.ID

    inspected, err := client.API().VolumeInspect(nil, "scaleio", created.ID, false)
    assert.Nil(t, err)
    assert.Equal(t, inspected.Name, volumeCreateRequest.Name)
    assert.Equal(t, inspected.Size, *volumeCreateRequest.Size)

  }



  deleteVolumeTest := func(config gofig.Config, client client.Client, t *testing.T) {

    err := client.API().VolumeRemove(nil, "scaleio", volumeID)
    assert.Nil(t, err)

    inspected, err := client.API().VolumeInspect(nil, "scaleio", volumeID, false)
    assert.Error(t, err)
    assert.Empty(t, inspected)
  }

  apitests.RunGroup(
    t, scaleio.Name, configYAML,
    volumeEndpointTest,
    createVolumeTest,
    deleteVolumeTest)
}

//////////////////////
///  Test Helpers  ///
//////////////////////

func getSdcLocalGUID() *types.InstanceID {
  // get sdc kernel guid
  // /bin/emc/scaleio/drv_cfg --query_guid
  // sdcKernelGuid := "271bad82-08ee-44f2-a2b1-7e2787c27be1"

  out, err := exec.Command("/opt/emc/scaleio/sdc/bin/drv_cfg", "--query_guid").Output()
  if err != nil {
    return &types.InstanceID{}
  }
  sdcGUID := strings.Replace(string(out), "\n", "", -1)
  return &types.InstanceID{
    ID:   sdcGUID,
  }
}
