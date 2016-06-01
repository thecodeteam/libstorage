package types

import "syscall"

const (
	// MountOptUnknown is an unknown option.
	MountOptUnknown = MountOption(0)

	// MountOptReadOnly will mount the file system read-only.
	MountOptReadOnly = MountOption(syscall.MS_RDONLY)

	// MountOptNoSUID will not allow set-user-identifier or set-group-identifier
	// bits to take effect.
	MountOptNoSUID = MountOption(syscall.MS_NOSUID)

	// MountOptNoDev will not interpret character or block special devices on
	// the file system.
	MountOptNoDev = MountOption(syscall.MS_NODEV)

	// MountOptNoExec will not allow execution of any binaries on the mounted
	// file system.
	MountOptNoExec = MountOption(syscall.MS_NOEXEC)

	// MountOptSync will allow I/O to the file system to be done synchronously.
	MountOptSync = MountOption(syscall.MS_SYNCHRONOUS)

	// MountOptDirSync will force all directory updates within the file system
	// to be done synchronously. This affects the following system calls:
	// create, link, unlink, symlink, mkdir, rmdir, mknod and rename.
	MountOptDirSync = MountOption(syscall.MS_DIRSYNC)

	// MountOptRemount will attempt to remount an already-mounted file system.
	// This is commonly used to change the mount flags for a file system,
	// especially to make a readonly file system writeable. It does not change
	// device or mount point.
	MountOptRemount = MountOption(syscall.MS_REMOUNT)

	// MountOptMandLock will force mandatory locks on a filesystem.
	MountOptMandLock = MountOption(syscall.MS_MANDLOCK)

	// MountOptNoATime will not update the file access time when reading from
	// a file.
	MountOptNoATime = MountOption(syscall.MS_NOATIME)

	// MountOptNoDirATime will not update the directory access time.
	MountOptNoDirATime = MountOption(syscall.MS_NODIRATIME)

	// MountOptBind remounts a subtree somewhere else.
	MountOptBind = MountOption(syscall.MS_BIND)

	// MountOptRBind remounts a subtree and all possible submounts somewhere
	// else.
	MountOptRBind = MountOption(syscall.MS_BIND | syscall.MS_REC)

	// MountOptUnbindable creates a mount which cannot be cloned through a bind
	// operation.
	MountOptUnbindable = MountOption(syscall.MS_UNBINDABLE)

	// MountOptRUnbindable marks the entire mount tree as UNBINDABLE.
	MountOptRUnbindable = MountOption(syscall.MS_UNBINDABLE | syscall.MS_REC)

	// MountOptPrivate creates a mount which carries no propagation abilities.
	MountOptPrivate = MountOption(syscall.MS_PRIVATE)

	// MountOptRPrivate marks the entire mount tree as PRIVATE.
	MountOptRPrivate = MountOption(syscall.MS_PRIVATE | syscall.MS_REC)

	// MountOptSlave creates a mount which receives propagation from its master,
	// but not vice versa.
	MountOptSlave = MountOption(syscall.MS_SLAVE)

	// MountOptRSlave marks the entire mount tree as SLAVE.
	MountOptRSlave = MountOption(syscall.MS_SLAVE | syscall.MS_REC)

	// MountOptShared creates a mount which provides the ability to create
	// mirrors of that mount such that mounts and unmounts within any of the
	// mirrors propagate to the other mirrors.
	MountOptShared = MountOption(syscall.MS_SHARED)

	// MountOptRShared marks the entire mount tree as SHARED.
	MountOptRShared = MountOption(syscall.MS_SHARED | syscall.MS_REC)

	// MountOptRelATime updates inode access times relative to modify or
	// change time.
	MountOptRelATime = MountOption(syscall.MS_RELATIME)

	// MountOptStrictATime allows to explicitly request full atime updates.
	// This makes it possible for the kernel to default to relatime or noatime
	// but still allow userspace to override it.
	MountOptStrictATime = MountOption(syscall.MS_STRICTATIME)
)

var (
	mountOptToStr = map[MountOption]string{
		MountOptReadOnly:    "ro",
		MountOptNoSUID:      "nosuid",
		MountOptNoDev:       "nodev",
		MountOptNoExec:      "noexec",
		MountOptSync:        "sync",
		MountOptDirSync:     "dirsync",
		MountOptRemount:     "remount",
		MountOptMandLock:    "mand",
		MountOptNoATime:     "noatime",
		MountOptNoDirATime:  "nodiratime",
		MountOptBind:        "bind",
		MountOptRBind:       "rbind",
		MountOptUnbindable:  "unbindable",
		MountOptRUnbindable: "runbindable",
		MountOptPrivate:     "private",
		MountOptRPrivate:    "rprivate",
		MountOptSlave:       "slave",
		MountOptRSlave:      "rslave",
		MountOptShared:      "shared",
		MountOptRShared:     "rshared",
		MountOptRelATime:    "relatime",
		MountOptStrictATime: "strictatime",
	}

	mountStrToOpt = map[string]MountOption{
		"ro":          MountOptReadOnly,
		"nosuid":      MountOptNoSUID,
		"nodev":       MountOptNoDev,
		"noexec":      MountOptNoExec,
		"sync":        MountOptSync,
		"dirsync":     MountOptDirSync,
		"remount":     MountOptRemount,
		"mand":        MountOptMandLock,
		"noatime":     MountOptNoATime,
		"nodiratime":  MountOptNoDirATime,
		"bind":        MountOptBind,
		"rbind":       MountOptRBind,
		"unbindable":  MountOptUnbindable,
		"runbindable": MountOptRUnbindable,
		"private":     MountOptPrivate,
		"rprivate":    MountOptRPrivate,
		"slave":       MountOptSlave,
		"rslave":      MountOptRSlave,
		"shared":      MountOptShared,
		"rshared":     MountOptRShared,
		"relatime":    MountOptRelATime,
		"strictatime": MountOptStrictATime,
	}
)
