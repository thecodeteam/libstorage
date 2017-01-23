// +build !libstorage_storage_driver libstorage_storage_driver_s3fs

package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/akutz/goof"
	"github.com/akutz/gotil"

	"github.com/codedellemc/libstorage/api/types"
	"github.com/codedellemc/libstorage/drivers/storage/s3fs"
)

// Supported returns eiter current platform supports s3fs or not
func Supported(ctx types.Context) (bool, error) {
	return gotil.FileExistsInPath(s3fs.CmdName), nil
}

// InstanceID returns the instance ID for the local host.
func InstanceID(ctx types.Context) (*types.InstanceID, error) {
	hostname := os.Getenv("S3FS_INSTANCE_ID")
	if hostname == "" {
		var err error
		hostname, err = os.Hostname()
		if err != nil {
			return nil, err
		}
	} else {
		ctx.Info("Use InstanceID from env " + hostname)
	}
	return &types.InstanceID{
		ID:     hostname,
		Driver: s3fs.Name,
	}, nil
}

// baseURI is base URI: s3fs://
var baseURI = fmt.Sprintf("%s://", s3fs.Name)

// IsS3FSURI checks if uri has requried prefix
func IsS3FSURI(uri string) bool {
	return strings.HasPrefix(uri, baseURI)
}

// BucketFromURI extracts bucket name from device URI
func BucketFromURI(uri string) string {
	return strings.TrimPrefix(uri, baseURI)
}

// BucketURI makes bucket URI in form s3fs://bucket
func BucketURI(bucket string) string {
	return fmt.Sprintf("%s%s", baseURI, bucket)
}

// FindBucket finds mounted bucket name by mount point
func FindBucket(
	ctx types.Context,
	mountPoint string) (string, bool) {

	if buckets, err := MountedBuckets(ctx); err == nil {
		for b, mp := range buckets {
			if mp == mountPoint {
				return b, true
			}
		}
	}
	return "", false
}

// FindMountPoint finds mount point by bucket name
func FindMountPoint(
	ctx types.Context,
	bucket string) (string, bool) {

	if buckets, err := MountedBuckets(ctx); err == nil {
		b, ok := buckets[bucket]
		return b, ok
	}
	return "", false
}

// MountedBuckets enumerates mounted bucket
// and returns a map of buckets to their mount points
func MountedBuckets(
	ctx types.Context) (map[string]string, error) {

	buckets := map[string]string{}
	command := exec.Command("bash", "-c",
		fmt.Sprintf("ps ax | awk '{if($5==\"%s\"){print($6\",\"$7)}}'",
			s3fs.CmdName))
	output, err := command.CombinedOutput()
	if err == nil {
		for _, text := range strings.Split(string(output), "\n") {
			if text != "" {
				pair := strings.Split(text, ",")
				buckets[pair[0]] = pair[1]
			}
		}
	} else {
		ctx.Warning(fmt.Sprintf(fmt.Sprintf(
			"Cant read s3fs processes: %s",
			string(output))))
	}
	ctx.Debugf("DBG: mounted buckets: %s", buckets)
	return buckets, nil
}

// Mount performs mounting via s3fs fuse cmd
func Mount(
	ctx types.Context,
	credFile, bucket, mountPoint string,
	opts *types.DeviceMountOpts) error {

	ctx.Debugf("DBG: s3fs mount bucket '%s' to mount point '%s'",
		bucket, mountPoint)

	// TODO: use opts
	command := exec.Command(
		s3fs.CmdName, bucket, mountPoint,
		"-o", fmt.Sprintf("passwd_file=%s", credFile))
	output, err := command.CombinedOutput()
	if err != nil {
		return goof.WithError(fmt.Sprintf(
			"failed to mount bucket %s, output '%s'",
			bucket, string(output)), err)
	}

	return nil
}
