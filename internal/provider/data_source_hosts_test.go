package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceHosts(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() {},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testPritunlHostsConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("num_hosts", "1"),
				),
			},
		},
	})
}

func testPritunlHostsConfig() string {
	return fmt.Sprintf(`
data "pritunl_hosts" "my-server-hosts" {}

output "num_hosts" {
  value = length(data.pritunl_hosts.my-server-hosts.hosts)
}
`)
}
