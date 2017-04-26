// +build linux
// +build !libstorage_storage_driver

package remote

import (
	// import to load
	_ "github.com/codedellemc/libstorage/drivers/storage/vmdk/executor"
)
