package types

import (
	"bytes"
	"fmt"
	"regexp"
)

/*
$ mount
/dev/mapper/mea--vg-root on / type ext4 (rw,errors=remount-ro)
proc on /proc type proc (rw,noexec,nosuid,nodev)
sysfs on /sys type sysfs (rw,noexec,nosuid,nodev)
none on /sys/fs/cgroup type tmpfs (rw)
none on /sys/fs/fuse/connections type fusectl (rw)
none on /sys/kernel/debug type debugfs (rw)
none on /sys/kernel/security type securityfs (rw)
udev on /dev type devtmpfs (rw,mode=0755)
devpts on /dev/pts type devpts (rw,noexec,nosuid,gid=5,mode=0620)
tmpfs on /run type tmpfs (rw,noexec,nosuid,size=10%,mode=0755)
none on /run/lock type tmpfs (rw,noexec,nosuid,nodev,size=5242880)
none on /run/shm type tmpfs (rw,nosuid,nodev)
none on /run/user type tmpfs (rw,noexec,nosuid,nodev,size=104857600,mode=0755)
none on /sys/fs/pstore type pstore (rw)
/dev/sda1 on /boot type ext2 (rw)
systemd on /sys/fs/cgroup/systemd type cgroup (rw,noexec,nosuid,nodev,none,name=systemd)
go on /media/sf_go type vboxsf (gid=999,rw)
/tmp/one on /tmp/one-bind type none (rw,bind)
*/

// MarshalText marshals the MountInfo object to its textual representation.
func (i *MountInfo) MarshalText() ([]byte, error) {
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "%s on %s type %s", i.DevicePath, i.MountPoint, i.FSType)
	if len(i.Opts) > 0 {
		fmt.Fprintf(buf, " (%s)", i.Opts)
	}
	return buf.Bytes(), nil
}

/*
mountInfoRX is the regex used for matching the output of the Linux mount cmd

$1 = devicePath
$2 = mountPoint
$3 = fileSystemType
$4 = mountOpts
*/
var mountInfoRX = regexp.MustCompile(`^(.+) on (.+) type (.+?)(?: \((.+)\))?$`)
