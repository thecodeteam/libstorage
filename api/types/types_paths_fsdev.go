package types

import "regexp"

// FileSystemDevicePath is a path to a filesystem device.
type FileSystemDevicePath string

// String returns the string representation of the file system device path.
func (p FileSystemDevicePath) String() string {
	return string(p)
}

var (
	nfsDevPathRX = regexp.MustCompile(`^([^:]+):(.+)$`)
)

// IsNFS returns information about a file system device path as if the path is
// an NFS export.
func (p FileSystemDevicePath) IsNFS() (
	ok bool,
	remoteHost string,
	remoteDir string) {

	m := nfsDevPathRX.FindStringSubmatch(string(p))
	if len(m) == 0 {
		return false, "", ""
	}

	return true, m[1], m[2]
}

// IsBind returns a flag indicating whether or not the path appears to be a
// bind mount path. This is decided based on whether or not the device path is
// in the /dev directory.
func (p FileSystemDevicePath) IsBind() bool {
	nfs, _, _ := p.IsNFS()
	return !nfs && p.isBind()
}
