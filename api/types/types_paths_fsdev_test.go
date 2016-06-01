package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileSystemDevicePathIsNFS(t *testing.T) {

	nfs, remoteHost, remoteDir := FileSystemDevicePath(
		"/dev/xvda").IsNFS()
	assert.False(t, nfs)

	nfs, remoteHost, remoteDir = FileSystemDevicePath(
		"server1:/shares/mine").IsNFS()
	assert.True(t, nfs)
	assert.Equal(t, "server1", remoteHost)
	assert.Equal(t, "/shares/mine", remoteDir)

	nfs, remoteHost, remoteDir = FileSystemDevicePath(
		"/home/myhome/share").IsNFS()
	assert.False(t, nfs)
}

func TestFileSystemDevicePathIsBind(t *testing.T) {

	assert.False(t, FileSystemDevicePath("/dev/xvda").IsBind())
	assert.False(t, FileSystemDevicePath("server1:/shares/mine").IsBind())
	assert.True(t, FileSystemDevicePath("/home/myhome/share").IsBind())
}

func TestFileSystemDevicePathMarshalJSON(t *testing.T) {

	buf, err := json.Marshal(FileSystemDevicePath(`/dev/xvda`))
	if err != nil {
		assert.NoError(t, err)
		t.FailNow()
	}
	assert.EqualValues(t, []byte(`"/dev/xvda"`), buf)

	buf, err = json.Marshal(FileSystemDevicePath(`server1:/shares/mine`))
	if err != nil {
		assert.NoError(t, err)
		t.FailNow()
	}
	assert.EqualValues(t, []byte(`"server1:/shares/mine"`), buf)

	buf, err = json.Marshal(FileSystemDevicePath(`/home/myhome/share`))
	if err != nil {
		assert.NoError(t, err)
		t.FailNow()
	}
	assert.EqualValues(t, []byte(`"/home/myhome/share"`), buf)
}
