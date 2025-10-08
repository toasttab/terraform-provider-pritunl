package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPritunlOrganization(t *testing.T) {

	t.Run("creates organizations without error", func(t *testing.T) {
		orgName := "tfacc-org1"

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("pritunl_organization.test", "name", orgName),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testPritunlOrganizationConfig(orgName),
					Check:  check,
				},
				// import test
				importStep("pritunl_organization.test"),
			},
		})
	})
}

func testPritunlOrganizationConfig(name string) string {
	return fmt.Sprintf(`
		resource "pritunl_organization" "test" {
			name    = "%[1]s"
		}
	`, name)
}
