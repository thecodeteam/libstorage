// +build !libstorage_storage_driver libstorage_storage_driver_DRIVERNAME

package DRIVERNAME

import (
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/codedellemc/libstorage/api/server"
	"github.com/codedellemc/libstorage/api/types"
)

// Put contents of sample config.yml here
var (
	configYAML = []byte(``)
)

var volumeName string
var volumeName2 string

// Check environment vars to see whether or not to run this test
func skipTests() bool {
	travis, _ := strconv.ParseBool(os.Getenv("TRAVIS"))
	noTest, _ := strconv.ParseBool(os.Getenv("TEST_SKIP_DRIVERNAME"))
	return travis || noTest
}

// Set volume names to first part of UUID before the -
func init() {
	uuid, _ := types.NewUUID()
	uuids := strings.Split(uuid.String(), "-")
	volumeName = uuids[0]
	uuid, _ = types.NewUUID()
	uuids = strings.Split(uuid.String(), "-")
	volumeName2 = uuids[0]
}

func TestMain(m *testing.M) {
	server.CloseOnAbort()
	ec := m.Run()
	os.Exit(ec)
}
