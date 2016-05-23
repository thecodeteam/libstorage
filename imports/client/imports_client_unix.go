// +build linux darwin

package client

import (
	// load the os drivers
	_ "github.com/emccode/libstorage/drivers/os/unix"
)
