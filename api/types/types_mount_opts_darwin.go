package types

const (

	// MountOptUnknown is an unknown option.
	MountOptUnknown = MountOption(0)

	// MountOptReadOnly will mount the file system read-only.
	MountOptReadOnly = MountOption(0x00000001)

	// MountOptNoSUID will not allow set-user-identifier or set-group-identifier
	// bits to take effect.
	MountOptNoSUID = MountOption(0x00000008)

	// MountOptNoDev will not interpret character or block special devices on
	// the file system.
	MountOptNoDev = MountOption(0x00000010)

	// MountOptNoExec will not allow execution of any binaries on the mounted
	// file system.
	MountOptNoExec = MountOption(0x00000004)

	// MountOptSync will allow I/O to the file system to be done synchronously.
	MountOptSync = MountOption(0x00000002)

	// MountOptNoATime will not update the file access time when reading from
	// a file.
	MountOptNoATime = MountOption(0x10000000)

	// MountOptLocal indicates the file system is stored locally.
	MountOptLocal = MountOption(0x00001000)

	// MountOptQuota indicates quotas are enabled on the file system.
	MountOptQuota = MountOption(0x00002000)

	// MountOptRootFS identifies the root file system.
	MountOptRootFS = MountOption(0x00004000)

	// MountOptDontBrowse indicates the file system is not appropriate path to
	// user data
	MountOptDontBrowse = MountOption(0x00100000)

	// MountOptIgnoreOwnership indicates ownership information on file system
	// objects will be ignored
	MountOptIgnoreOwnership = MountOption(0x00200000)

	// MountOptAutoMounted indicates file system was mounted by auto mounter
	MountOptAutoMounted = MountOption(0x00400000)

	// MountOptJournaled indicates file system is journaled
	MountOptJournaled = MountOption(0x00800000)

	// MountOptNoUserXattr indicates user extended attributes are not allowed
	MountOptNoUserXattr = MountOption(0x01000000)

	// MountOptDefWrite indicates the file system should defer writes
	MountOptDefWrite = MountOption(0x02000000)

	// MountOptMultiLabel indicates MAC support for individual labels
	MountOptMultiLabel = MountOption(0x04000000)
)

var (
	mountOptToStr = map[MountOption]string{
		MountOptReadOnly:        "read-only",
		MountOptNoSUID:          "nosuid",
		MountOptNoDev:           "nodev",
		MountOptNoExec:          "noexec",
		MountOptSync:            "sync",
		MountOptNoATime:         "noatime",
		MountOptLocal:           "local",
		MountOptQuota:           "quota",
		MountOptRootFS:          "rootfs",
		MountOptDontBrowse:      "nobrowse",
		MountOptIgnoreOwnership: "noowners",
		MountOptAutoMounted:     "automounted",
		MountOptJournaled:       "journaled",
		MountOptNoUserXattr:     "nouserxattr",
		MountOptDefWrite:        "defwrite",
	}

	mountStrToOpt = map[string]MountOption{
		"read-only":   MountOptReadOnly,
		"nosuid":      MountOptNoSUID,
		"nodev":       MountOptNoDev,
		"noexec":      MountOptNoExec,
		"sync":        MountOptSync,
		"noatime":     MountOptNoATime,
		"local":       MountOptLocal,
		"quota":       MountOptQuota,
		"rootfs":      MountOptRootFS,
		"nobrowse":    MountOptDontBrowse,
		"noowners":    MountOptIgnoreOwnership,
		"automounted": MountOptAutoMounted,
		"journaled":   MountOptJournaled,
		"nouserxattr": MountOptNoUserXattr,
		"defwrite":    MountOptDefWrite,
	}
)
