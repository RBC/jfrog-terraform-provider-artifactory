package security_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccApiKey(t *testing.T) {
	client := acctest.GetTestResty(t)
	version, err := util.GetArtifactoryVersion(client)
	if err != nil {
		t.Fatal(err)
	}
	valid, err := util.CheckVersion(version, "7.98.1")
	if err != nil {
		t.Fatal(err)
	}
	if valid {
		t.Skipf("Artifactory version %s is 7.98.1 or later", version)
	}

	fqrn := "artifactory_api_key.foobar"
	const apiKey = `
		resource "artifactory_api_key" "foobar" {}
	`

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckApiKeyDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: apiKey,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(fqrn, "api_key"),
				),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckApiKeyDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(util.ProviderMetadata).Client
		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		data := security.ApiKey{}

		_, err := client.R().SetResult(&data).Get(security.ApiKeyEndpoint)
		if err != nil {
			return err
		}

		if data.ApiKey != "" {
			return fmt.Errorf("error: API key %s still exists", rs.Primary.ID)
		}
		return nil
	}
}
