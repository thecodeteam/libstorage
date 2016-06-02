// +build darwin

package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMountInfoParse(t *testing.T) {
	expected := &MountInfo{
		DevicePath: FileSystemDevicePath("/dev/disk1"),
		MountPoint: "/",
		Opts:       MountOptions{MountOptLocal, MountOptJournaled},
		FSType:     "hfs",
	}
	actual := ParseMountInfo("/dev/disk1 on / (hfs, local, journaled)")
	assert.Equal(t, expected, actual)
	assert.Equal(t, "/dev/disk1 on / (hfs, local, journaled)", actual.String())
}
