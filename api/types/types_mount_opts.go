package types

import (
	"bytes"
	"encoding/json"
	"runtime"
	"strings"

	"github.com/akutz/goof"
)

// MountOption is a mount option.
type MountOption int

// MountOptions are a mount options string.
type MountOptions []MountOption

// String returns the string representation of the MountOption.
func (o MountOption) String() string {
	if buf, err := o.MarshalText(); err == nil {
		return string(buf)
	}
	return ""
}

func (o MountOption) bytes() []byte {
	if v, ok := mountOptToStr[o]; ok {
		return []byte(v)
	}
	return nil
}

// ParseMountOption parses a mount option.
func ParseMountOption(text string) MountOption {
	o := MountOptUnknown
	o.UnmarshalText([]byte(text))
	return o
}

// MarshalText marshals the MountOption to its string representation.
func (o MountOption) MarshalText() ([]byte, error) {
	if buf := o.bytes(); buf != nil {
		return buf, nil
	}
	return nil, goof.WithField("opt", int(o), "invalid mount option")
}

// UnmarshalText marshals the MountOption from its string representation.
func (o *MountOption) UnmarshalText(data []byte) error {
	text := string(data)
	if v, ok := mountStrToOpt[strings.ToLower(text)]; ok {
		*o = v
		return nil
	}
	return goof.WithField("opt", text, "invalid mount option")
}

const (
	commaByteVal byte = 44
	spaceByteVal byte = 32
)

var (
	commaSepBuf      = []byte{commaByteVal}
	commaSpaceSepBuf = []byte{commaByteVal, spaceByteVal}
)

// ParseMountOptions parses a mount options string.
func ParseMountOptions(text string) MountOptions {
	var opts MountOptions
	if err := opts.UnmarshalText([]byte(text)); err == nil {
		return opts
	}
	return nil
}

// String returns the string representation of the MountOptions object.
func (opts MountOptions) String() string {
	if s, err := opts.MarshalText(); err == nil {
		return string(s)
	}
	return ""
}

// MarshalText marshals the MountOptions to its string representation.
func (opts MountOptions) MarshalText() ([]byte, error) {
	buf := &bytes.Buffer{}
	for x, o := range opts {
		if v, ok := mountOptToStr[o]; ok {
			buf.WriteString(v)
			if x < (len(opts) - 1) {
				switch runtime.GOOS {
				case "linux":
					buf.WriteString(",")
				case "darwin":
					buf.WriteString(", ")
				}
			}
		}
	}
	return buf.Bytes(), nil
}

// UnmarshalText marshals the MountOptions from its string representation.
func (opts *MountOptions) UnmarshalText(text []byte) error {
	var sepBuf []byte
	switch runtime.GOOS {
	case "linux":
		sepBuf = commaSepBuf
	case "darwin":
		sepBuf = commaSpaceSepBuf
	}
	optBufs := bytes.Split(text, sepBuf)
	for _, optText := range optBufs {
		if o := ParseMountOption(string(optText)); o != MountOptUnknown {
			*opts = append(*opts, o)
		}
	}
	return nil
}

// MarshalJSON marshals the MountOptions to its JSON representation.
func (opts MountOptions) MarshalJSON() ([]byte, error) {
	strOpts := make([]string, len(opts))
	for i, o := range opts {
		strOpts[i] = o.String()
	}
	return json.Marshal(strOpts)
}

// UnmarshalJSON marshals the MountOptions from its JSON representation.
func (opts *MountOptions) UnmarshalJSON(text []byte) error {
	strOpts := []string{}
	if err := json.Unmarshal(text, &strOpts); err != nil {
		return err
	}
	for _, optText := range strOpts {
		if o := ParseMountOption(optText); o != MountOptUnknown {
			*opts = append(*opts, o)
		}
	}
	return nil
}
