// +build !libstorage_storage_driver libstorage_storage_driver_DRIVERNAME

package template

import gofigCore "github.com/akutz/gofig"

const (
	// Name is the provider's name.
	Name = "DRIVERNAME"
)

func init() {
	r := gofigCore.NewRegistration("DRIVERNAME")

	gofigCore.Register(r)
}
