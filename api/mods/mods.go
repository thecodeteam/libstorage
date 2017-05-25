// +build go1.8,linux

package mods

import (
	"io/ioutil"
	"os"
	"path"
	"plugin"
	"strconv"
	"strings"
	"sync"

	"github.com/akutz/gotil"

	"github.com/codedellemc/libstorage/api/types"
)

var (
	loadedMods     = map[string]bool{}
	loadedModsLock = sync.Mutex{}
)

// LoadModules loads the shared objects present on the file system
// as libStorage plug-ins.
func LoadModules(
	ctx types.Context,
	pathConfig *types.PathConfig) {

	disabled, _ := strconv.ParseBool(
		os.Getenv("LIBSTORAGE_PLUGINS_DISABLED"))
	if disabled {
		ctx.Debug("plugin support disabled")
		return
	}

	loadedModsLock.Lock()
	defer loadedModsLock.Unlock()

	if !gotil.FileExists(pathConfig.Mod) {
		return
	}
	modFiles, err := ioutil.ReadDir(pathConfig.Mod)
	if err != nil {
		ctx.WithField("path", pathConfig.Mod).Warn(
			"failed to list module files")
		return
	}
	for _, f := range modFiles {
		modFilePath := f.Name()
		modFilePath = path.Join(pathConfig.Mod, modFilePath)
		ctx.WithField(
			"path", modFilePath).Debug(
			"loading module")
		lcModFilePath := strings.ToLower(modFilePath)
		if loaded, ok := loadedMods[lcModFilePath]; ok && loaded {
			ctx.WithField(
				"path", modFilePath).Debug(
				"already loaded")
			continue
		}
		_, err := plugin.Open(modFilePath)
		if err != nil {
			ctx.WithError(err).WithField(
				"path", modFilePath).Error(
				"error opening module")
			continue
		}
		loadedMods[lcModFilePath] = true
		ctx.WithField(
			"path", modFilePath).Info(
			"loaded module")
	}
}
