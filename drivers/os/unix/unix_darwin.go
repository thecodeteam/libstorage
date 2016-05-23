// +build darwin

package unix

//#cgo CFLAGS: -I${SRCDIR}
//#include "unix_darwin.h"
import "C"
import (
	"unsafe"

	"github.com/akutz/goof"

	"github.com/emccode/libstorage/api/types"
)

const (
	// ReadOnly will mount the file system read-only.
	ReadOnly = C.MNT_RDONLY

	// NoSetUserID will not allow set-user-identifier or set-group-identifier
	// bits to take effect.
	NoSetUserID = C.MNT_NOSUID

	// NoDev will not interpret character or block special devices on the file
	// system.
	NoDev = C.MNT_NODEV

	// NoExec will not allow execution of any binaries on the mounted file
	// system.
	NoExec = C.MNT_NOEXEC

	// Synchronous will allow I/O to the file system to be done synchronously.
	Synchronous = C.MNT_SYNCHRONOUS

	// NoAccessTime will not update the file access time when reading from a
	// file.
	NoAccessTime = C.MNT_NOATIME

	// Wait instructs calls to get information about a filesystem to refresh
	// information about a filesystem before returning it, causing the call to
	// block until the refresh operation is complete.
	Wait = C.MNT_WAIT

	// NoWait instructs calls to get information about a filesystem to return
	// any available information immediately without waiting.
	NoWait = C.MNT_NOWAIT
)

type fsInfo struct {
	blockSize          int64
	ioSize             int64
	blocks             int64
	blocksFree         int64
	blocksAvail        int64
	files              int64
	filesFree          int64
	fileSystemTypeID   int8
	fileSystemTypeName string
	mountPath          string
	devicePath         string
	mountFlags         int64
}

func statFS(mountPoint string) (*fsInfo, error) {

	r := C._statfs(C.CString(mountPoint))
	if r.val != nil {
		defer C.free(unsafe.Pointer(r.val))
	}

	if r.err != 0 {
		return nil, goof.WithFields(goof.Fields{
			"mountPoint": mountPoint,
			"error":      r.err,
		}, "statFS error")
	}

	return toFSInfoFromStatFS(r.val), nil
}

func mounts(
	ctx types.Context,
	deviceName, mountPoint string,
	opts types.Store) ([]*types.MountInfo, error) {

	fsInfo, err := getMountInfo(ctx, true)
	if err != nil {
		return nil, err
	}

	return toMountInfoArray(fsInfo), nil
}

func mount(device, target, mType, options string) error {

	return nil
}

func format(
	ctx types.Context,
	deviceName string,
	opts *types.DeviceFormatOpts) error {

	return nil
}

func getMountInfo(ctx types.Context, wait bool) ([]*fsInfo, error) {

	var flags int
	if wait {
		flags = Wait
	} else {
		flags = NoWait
	}

	r := C._getmntinfo(C.int(flags))
	if r.err != 0 {
		return nil, goof.WithFields(goof.Fields{
			"wait":  wait,
			"flags": flags,
			"len":   r.len,
			"error": r.err,
		}, "getMountInfo error")
	}

	ctx.WithField("len", r.len).Debug("got mount info")
	fsiList := make([]*fsInfo, r.len)
	miSlice := (*[1 << 30]C.struct_statfs)(unsafe.Pointer(r.val))[:r.len:r.len]

	for x, mi := range miSlice {
		fsiList[x] = toFSInfoFromStatFS(&mi)
	}

	return fsiList, nil
}

func toMountInfoArray(val []*fsInfo) []*types.MountInfo {

	newVal := make([]*types.MountInfo, len(val))
	for x, fsi := range val {
		newVal[x] = toMountInfo(fsi)
	}
	return newVal
}

func toMountInfo(val *fsInfo) *types.MountInfo {

	return &types.MountInfo{
		Source:     val.devicePath,
		MountPoint: val.mountPath,
		FSType:     val.fileSystemTypeName,
	}
}

func toFSInfoFromStatFS(val *C.struct_statfs) *fsInfo {

	return &fsInfo{
		blockSize:          int64(val.f_bsize),
		ioSize:             int64(val.f_iosize),
		blocks:             int64(val.f_blocks),
		blocksFree:         int64(val.f_bfree),
		blocksAvail:        int64(val.f_bavail),
		files:              int64(val.f_files),
		filesFree:          int64(val.f_ffree),
		fileSystemTypeID:   int8(val.f_type),
		fileSystemTypeName: C.GoString(&val.f_fstypename[0]),
		mountPath:          C.GoString(&val.f_mntonname[0]),
		devicePath:         C.GoString(&val.f_mntfromname[0]),
		mountFlags:         int64(val.f_flags),
	}
}
