// +build go1.8,linux,mods

package mods

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"plugin"
	"strconv"
	"sync"

	"github.com/akutz/gotil"

	"github.com/codedellemc/libstorage/api/registry"
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
	pathConfig *types.PathConfig) error {

	disabled, _ := strconv.ParseBool(
		os.Getenv("LIBSTORAGE_PLUGINS_DISABLED"))
	if disabled {
		ctx.Debug("plugin support disabled")
		return nil
	}

	loadedModsLock.Lock()
	defer loadedModsLock.Unlock()

	if !gotil.FileExists(pathConfig.Mod) {
		return fmt.Errorf("error: invalid mod dir: %v", pathConfig.Mod)
	}

	modFilePathMatches, err := filepath.Glob(path.Join(pathConfig.Mod, "/*.so"))
	if err != nil {
		// since the only possible error is ErrBadPattern then make sure
		// it panics the program since it should never be an invalid pattern
		panic(err)
	}

	for _, modFilePath := range modFilePathMatches {
		ctx.WithField(
			"path", modFilePath).Debug(
			"loading module")

		if loaded, ok := loadedMods[modFilePath]; ok && loaded {
			ctx.WithField(
				"path", modFilePath).Debug(
				"already loaded")
			continue
		}

		p, err := plugin.Open(modFilePath)
		if err != nil {
			ctx.WithError(err).WithField(
				"path", modFilePath).Error(
				"error opening module")
			continue
		}

		if err := loadPluginTypes(ctx, p); err != nil {
			continue
		}

		loadedMods[modFilePath] = true
		ctx.WithField(
			"path", modFilePath).Info(
			"loaded module")
	}

	return nil
}

func loadPluginTypes(ctx types.Context, p *plugin.Plugin) error {
	// lookup the plug-in's Types symbol; it's the type map used to
	// register the plug-in's modules
	tmapObj, err := p.Lookup("Types")
	if err != nil {
		ctx.WithError(err).Error("error looking up type map")
		return err
	}

	// assert that the Types symbol is a *map[string]func() interface{}
	tmapPtr, tmapOk := tmapObj.(*map[string]func() interface{})
	if !tmapOk {
		err := fmt.Errorf("invalid type map: %T", tmapObj)
		ctx.Error(err)
		return err
	}

	// assert that the type map pointer is not nil
	if tmapPtr == nil {
		err := fmt.Errorf("nil type map: type=%[1]T val=%[1]v", tmapPtr)
		ctx.Error(err)
		return err
	}

	// dereference the type map pointer
	tmap := *tmapPtr

	// register the plug-in's modules
	for k, v := range tmap {
		registry.RegisterModType(k, v)
	}

	return nil
}
