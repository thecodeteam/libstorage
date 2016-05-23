// +build darwin

package unix

import (
	"testing"

	"github.com/akutz/goof"
	"github.com/stretchr/testify/assert"

	"github.com/emccode/libstorage/api/context"
	"github.com/emccode/libstorage/api/utils"
)

func init() {
	goof.IncludeFieldsInError = true
	goof.IncludeFieldsInString = true
	goof.IncludeFieldsInFormat = true
}

func TestMounts(t *testing.T) {

	ctx := context.Background()
	store := utils.NewStore()
	d := newDriver()

	mounts, err := d.Mounts(ctx, "", "", store)
	assert.NoError(t, err)
	assert.True(t, len(mounts) > 1)

	mounts, err = d.Mounts(ctx, "", "/", store)
	assert.NoError(t, err)
	assert.True(t, len(mounts) == 1)
}

func TestIsMounted(t *testing.T) {

	ctx := context.Background()
	store := utils.NewStore()
	d := newDriver()

	isMounted, err := d.IsMounted(ctx, "/", store)
	assert.NoError(t, err)
	assert.True(t, isMounted)
}

func TestStatFS(t *testing.T) {

	r, err := statFS("/")
	assert.NoError(t, err)
	t.Logf("%+v", r)
}

func TestGetMountInfoAndStatFS(t *testing.T) {

	r, err := getMountInfo(context.Background(), true)
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}

	for _, fsi := range r {
		statFSResult, err := statFS(fsi.mountPath)
		assert.NoError(t, err)
		t.Logf("%+v", statFSResult)
	}
}
