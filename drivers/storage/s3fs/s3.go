// +build !libstorage_storage_driver libstorage_storage_driver_s3fs

package s3fs

import (
	gofigCore "github.com/akutz/gofig"
	gofig "github.com/akutz/gofig/types"
)

const (
	// Name is the provider's name.
	Name = "s3fs"

	// CmdName of the s3fs cmd utility
	CmdName = "s3fs"

	// TagDelimiter separates tags from volume or snapshot names
	TagDelimiter = "/"

	// BucketsKey is a name of config parameter with buckets list
	BucketsKey = "buckets"

	// CredFilePathKey is a name of config parameter with cred file path
	CredFilePathKey = "cred_file"

	// TagKey is a tag key
	TagKey = "tag"
)

const (
	// ConfigS3FS is a config key
	ConfigS3FS = Name

	// ConfigS3FSBucketsKey is a key for available buckets list
	ConfigS3FSBucketsKey = ConfigS3FS + "." + BucketsKey

	// ConfigS3FSCredFilePathKey is a key for cred file path
	ConfigS3FSCredFilePathKey = ConfigS3FS + "." + CredFilePathKey

	// ConfigS3FSTagKey is a config key
	ConfigS3FSTagKey = ConfigS3FS + "." + TagKey
)

func init() {
	r := gofigCore.NewRegistration("S3FS")
	r.Key(gofig.String, "", "",
		"List of buckets available as file systems",
		ConfigS3FSBucketsKey)
	r.Key(gofig.String, "", "",
		"File path with S3 credentials in format ID:KEY",
		ConfigS3FSCredFilePathKey)
	r.Key(gofig.String, "", "",
		"Tag prefix for S3FS naming",
		ConfigS3FSTagKey)
	gofigCore.Register(r)
}
