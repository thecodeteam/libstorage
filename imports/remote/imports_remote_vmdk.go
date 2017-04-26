// +build linux
// +build libstorage_storage_driver,libstorage_storage_driver_vmdk

package remote

import (
	// load the packages
	_ "github.com/codedellemc/libstorage/drivers/storage/vmdk/storage"
)
