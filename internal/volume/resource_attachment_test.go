package volume_test

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"testing"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/volume"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestVolumeAssignmentResource_Basic(t *testing.T) {
	var s hcloud.Server
	var s2 hcloud.Server
	var v hcloud.Volume
	tmplMan := testtemplate.Manager{}
	resServer := &server.RData{
		Name:  "vol-attachment",
		Type:  "cx11",
		Image: "ubuntu-20.04",
		Labels: map[string]string{
			"tf-test": fmt.Sprintf("tf-test-vol-attachment-%d", tmplMan.RandInt),
		},
		LocationName: "fsn1",
	}
	resServer.SetRName("server_attachment")

	resServer2 := &server.RData{
		Name:  "vol-attachment-2",
		Type:  "cx11",
		Image: "ubuntu-20.04",
		Labels: map[string]string{
			"tf-test": fmt.Sprintf("tf-test-vol-attachment-%d", tmplMan.RandInt),
		},
		LocationName: "fsn1",
	}
	resServer2.SetRName("server2_attachment")

	resVolume := &volume.RData{
		Name:         "volume-attachment",
		Size:         10,
		LocationName: "fsn1",
	}
	resVolume.SetRName("volume-attachment")

	res := &volume.RDataAttachment{
		VolumeID: resVolume.TFID() + ".id",
		ServerID: resServer.TFID() + ".id",
	}

	resMove := &volume.RDataAttachment{
		VolumeID: resVolume.TFID() + ".id",
		ServerID: resServer2.TFID() + ".id",
	}
	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, &s)),
		Steps: []resource.TestStep{
			{
				// Create a new Volume attachment using the required values
				// only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", resServer,
					"testdata/r/hcloud_server", resServer2,
					"testdata/r/hcloud_volume", resVolume,
					"testdata/r/hcloud_volume_attachment", res,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(resServer.TFID(), server.ByID(t, &s)),
					testsupport.CheckResourceExists(resVolume.TFID(), volume.ByID(t, &v)),
				),
			},
			{
				// Try to import the newly created volume attachment
				ResourceName:      res.TFID(),
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return fmt.Sprintf("%d", v.ID), nil
				},
			},
			{
				// Move the Volume to another server using the
				// attachment.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", resServer,
					"testdata/r/hcloud_server", resServer2,
					"testdata/r/hcloud_volume", resVolume,
					"testdata/r/hcloud_volume_attachment", resMove,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(resServer2.TFID(), server.ByID(t, &s2)),
					testsupport.CheckResourceExists(resVolume.TFID(), volume.ByID(t, &v)),
				),
			},
		},
	})
}
