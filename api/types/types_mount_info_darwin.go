package types

import (
	"bytes"
	"fmt"
	"regexp"
)

/*
$ mount
/dev/disk1 on / (hfs, local, journaled)
devfs on /dev (devfs, local, nobrowse)
map -hosts on /net (autofs, nosuid, automounted, nobrowse)
map auto_home on /home (autofs, automounted, nobrowse)
/tmp/one on /private/tmp/bind-one (osxfusefs, nodev, nosuid, synchronous, mounted by akutz)
bindfs@osxfuse1 on /private/tmp/bind-two (osxfusefs, nodev, nosuid, read-only, synchronous, mounted by akutz)
/dev/disk2s1 on /Volumes/VirtualBox (hfs, local, nodev, nosuid, read-only, noowners, quarantine, mounted by akutz)
*/

// MarshalText marshals the MountInfo object to its textual representation.
func (i *MountInfo) MarshalText() ([]byte, error) {
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "%s on %s (%s", i.DevicePath, i.MountPoint, i.FSType)
	if len(i.Opts) == 0 {
		fmt.Fprint(buf, ")")
	} else {
		fmt.Fprintf(buf, ", %s)", i.Opts)
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
var mountInfoRX = regexp.MustCompile(`^(.+) on (.+) \((.+?)(?:, (.+))\)$`)
