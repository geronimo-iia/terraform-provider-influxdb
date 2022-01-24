package influxdb

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/influxdata/influxdb/client"
)

func TestAccInfluxDBDatabase_basic(t *testing.T) {

	resourceName := "influxdb_database.test"
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "terraform-test"),
					resource.TestCheckResourceAttr(resourceName, "retention_policies.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccInfluxDBDatabase_retention(t *testing.T) {

	resourceName := "influxdb_database.test"
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseWithRPSConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "terraform-rp-test"),
					resource.TestCheckResourceAttr(resourceName, "retention_policies.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "retention_policies.*", map[string]string{
						"name":     "1day",
						"duration": "24h0m0s",
						"default":  "true",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDatabaseWithRPSUpdateConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(resourceName),
					resource.TestCheckResourceAttr("influxdb_database.test", "name", "terraform-rp-test"),
					resource.TestCheckResourceAttr(resourceName, "retention_policies.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "retention_policies.*", map[string]string{
						"name":     "2days",
						"duration": "48h0m0s",
						"default":  "false",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "retention_policies.*", map[string]string{
						"name":     "12weeks",
						"duration": "2016h0m0s",
						"default":  "true",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "retention_policies.*", map[string]string{
						"name":               "1week",
						"duration":           "168h0m0s",
						"default":            "false",
						"shardgroupduration": "2h0m0s",
					}),
				),
			},
		},
	})
}

func testAccCheckDatabaseExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No database id set")
		}

		conn := testAccProvider.Meta().(*client.Client)

		query := client.Query{
			Command: "SHOW DATABASES",
		}

		resp, err := conn.Query(query)
		if err != nil {
			return err
		}

		if resp.Err != nil {
			return resp.Err
		}

		for _, result := range resp.Results[0].Series[0].Values {
			if result[0] == rs.Primary.Attributes["name"] {
				return nil
			}
		}

		return fmt.Errorf("Database %q does not exist", rs.Primary.Attributes["name"])
	}
}

var testAccDatabaseConfig = `

resource "influxdb_database" "test" {
    name = "terraform-test"
}

`

var testAccDatabaseWithRPSConfig = `
resource "influxdb_database" "test" {
	name = "terraform-rp-test"
	retention_policies {
		name = "1day"
		duration = "24h0m0s"
		default = "true"
	}
}
`

var testAccDatabaseWithRPSUpdateConfig = `
resource "influxdb_database" "test" {
	name = "terraform-rp-test"
  
	retention_policies {
	  name     = "2days"
	  duration = "48h0m0s"
	}
  
	retention_policies {
	  name     = "12weeks"
	  duration = "2016h0m0s"
	  default  = "true"
	}
   
	retention_policies {
	  name               = "1week"
	  duration           = "168h0m0s"
	  shardgroupduration = "2h0m0s"
	}
  }
`
