package openstack

import (
	gofigCore "github.com/akutz/gofig"
	gofig "github.com/akutz/gofig/types"
)

const (
	// Name is the provider's name.
	Name = "openstack"
)

func init() {
	registerConfig()
}

func registerConfig() {
	r := gofigCore.NewRegistration("OpenStack")
	r.Key(gofig.String, "", "", "", "openstack.authURL")
	r.Key(gofig.String, "", "", "", "openstack.userID")
	r.Key(gofig.String, "", "", "", "openstack.userName")
	r.Key(gofig.String, "", "", "", "openstack.password")
	r.Key(gofig.String, "", "", "", "openstack.tenantID")
	r.Key(gofig.String, "", "", "", "openstack.tenantName")
	r.Key(gofig.String, "", "", "", "openstack.domainID")
	r.Key(gofig.String, "", "", "", "openstack.domainName")
	r.Key(gofig.String, "", "", "", "openstack.regionName")
	r.Key(gofig.String, "", "", "", "openstack.availabilityZoneName")
	gofigCore.Register(r)
}
