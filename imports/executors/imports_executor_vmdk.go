// +build linux
// +build libstorage_storage_executor,libstorage_storage_executor_vmdk

package executors

import (
	// load the packages
	_ "github.com/codedellemc/libstorage/drivers/storage/vmdk/executor"
)
