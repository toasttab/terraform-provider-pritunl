package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceHost(t *testing.T) {
	// pritunl.local sets in Makefile's "test" target
	existsHostname := "pritunl.local"
	notExistHostname := "not-exist-hostname"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() {},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testPritunlHostSimpleConfig(existsHostname),
				Check:  resource.ComposeTestCheckFunc(),
			},
			{
				Config:      testPritunlHostSimpleConfig(notExistHostname),
				ExpectError: regexp.MustCompile(fmt.Sprintf("Unable to get host with hostname %s", notExistHostname)),
			},
		},
	})
}

func testPritunlHostSimpleConfig(name string) string {
	return fmt.Sprintf(`
data "pritunl_host" "test" {
	hostname    = "%[1]s"
}
`, name)
}
