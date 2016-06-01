// +build linux darwin

package types

import (
	"regexp"
)

var (
	bindDevPathRX = regexp.MustCompile(`^/dev/.+$`)
)

// isBind returns a flag indicating whether or not the path appears to be a
// bind mount path. This is decided based on whether or not the device path is
// in the /dev directory.
func (p FileSystemDevicePath) isBind() bool {
	return !bindDevPathRX.MatchString(string(p))
}
