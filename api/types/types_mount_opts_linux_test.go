// +build linux

package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMountOptionParse(t *testing.T) {
	assert.Equal(t, MountOptReadOnly, ParseMountOption("ro"))
	assert.Equal(t, MountOptNoSUID, ParseMountOption("nosuid"))
	assert.Equal(t, MountOptNoExec, ParseMountOption("noexec"))
}

func TestMountOptionsParse(t *testing.T) {
	exepctedOpts := MountOptions{
		MountOptReadOnly,
		MountOptNoSUID,
		MountOptNoDev,
		MountOptNoExec,
	}
	opts := ParseMountOptions("ro,nosuid,nodev,noexec,bindfs")
	assert.Equal(t, exepctedOpts, opts)

	assert.Nil(t, ParseMountOptions(""))
}

func TestMountOptionsMarshalText(t *testing.T) {
	opts := MountOptions{
		MountOptReadOnly,
		MountOptNoSUID,
		MountOptNoDev,
		MountOptNoExec,
	}
	assert.Equal(t, "ro,nosuid,nodev,noexec", opts.String())
}

func TestMountOptionsUnmarshalText(t *testing.T) {
	var opts MountOptions
	err := opts.UnmarshalText(
		[]byte("ro,nosuid,nodev,noexec,bindfs"))
	if err != nil {
		assert.NoError(t, err)
		t.FailNow()
	}
	exepctedOpts := MountOptions{
		MountOptReadOnly,
		MountOptNoSUID,
		MountOptNoDev,
		MountOptNoExec,
	}
	assert.Equal(t, exepctedOpts, opts)
}

func TestMountOptionsMarshalJSON(t *testing.T) {
	opts := MountOptions{
		MountOptReadOnly,
		MountOptNoSUID,
		MountOptNoDev,
		MountOptNoExec,
	}
	buf, err := opts.MarshalJSON()
	if err != nil {
		assert.NoError(t, err)
		t.FailNow()
	}
	assert.Equal(t, `["ro","nosuid","nodev","noexec"]`, string(buf))
}

func TestMountOptionsUnmarshalJSON(t *testing.T) {
	var opts MountOptions
	err := opts.UnmarshalJSON([]byte(`[
  "ro",
  "nosuid",
  "bindfs",
  "nodev",
  "noexec"
]`))
	if err != nil {
		assert.NoError(t, err)
		t.FailNow()
	}
	exepctedOpts := MountOptions{
		MountOptReadOnly,
		MountOptNoSUID,
		MountOptNoDev,
		MountOptNoExec,
	}
	assert.Equal(t, exepctedOpts, opts)
}
