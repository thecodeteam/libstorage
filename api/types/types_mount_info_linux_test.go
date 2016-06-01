// +build linux

package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMountInfoParse(t *testing.T) {
	expected := &MountInfo{
		DevicePath: FileSystemDevicePath("proc"),
		MountPoint: "/proc",
		Opts:       MountOptions{MountOptNoExec, MountOptNoSUID, MountOptNoDev},
		FSType:     "proc",
	}
	actual := ParseMountInfo("proc on /proc type proc (rw,noexec,nosuid,nodev)")
	assert.Equal(t, expected, actual)
	assert.Equal(
		t, "proc on /proc type proc (noexec,nosuid,nodev)", actual.String())
}
