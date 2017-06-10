// +build !go1.8 !linux

package mods

import (
	"github.com/codedellemc/libstorage/api/types"
)

// LoadModules loads the shared objects present on the file system
// as libStorage plug-ins.
func LoadModules(
	ctx types.Context,
	pathConfig *types.PathConfig) {

	// NOOP
}
