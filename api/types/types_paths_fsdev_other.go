// +build !linux,!darwin

package types

import (
	"fmt"
	"runtime"
)

// isBind returns a flag indicating whether or not the path appears to be a
// bind mount path.
func (p FileSystemDevicePath) isBind() bool {
	panic(fmt.Errorf(
		"FileSystemDevicePath.IsBind unsupported on %s", runtime.GOOS))
}
