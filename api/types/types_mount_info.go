package types

// MountInfo reveals information about a particular mounted filesystem. This
// struct is populated from the content in the /proc/<pid>/mountinfo file.
type MountInfo struct {

	// DevicePath is the path of the mounted path.
	DevicePath FileSystemDevicePath `json:"devicePath"`

	// MountPoint indicates the mount point relative to the process's root.
	MountPoint string `json:"mountPoint"`

	// FSType indicates the type of filesystem, such as EXT3.
	FSType string `json:"fsType"`

	// Opts represents mount-specific options.
	Opts MountOptions `json:"opts"`
}

// MarshalText marshals the MountInfo object to its textual representation.
func (i *MountInfo) String() string {
	if s, err := i.MarshalText(); err == nil {
		return string(s)
	}
	return ""
}

// ParseMountInfo parses mount information.
func ParseMountInfo(text string) *MountInfo {
	i := &MountInfo{}
	i.UnmarshalText([]byte(text))
	return i
}

// UnmarshalText marshals the MountInfo from its textual representation.
func (i *MountInfo) UnmarshalText(data []byte) error {

	m := mountInfoRX.FindSubmatch(data)
	if len(m) == 0 {
		return nil
	}

	i.DevicePath = FileSystemDevicePath(m[1])
	i.MountPoint = string(m[2])
	i.FSType = string(m[3])
	i.Opts.UnmarshalText(m[4])

	return nil
}
