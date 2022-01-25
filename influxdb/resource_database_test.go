package influxdb

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/influxdata/influxdb/client"
)

func TestAccInfluxDBDatabase_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resourceName := "influxdb_database.test"
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resourceName := "influxdb_database.test"
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseWithRPSConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "retention_policies.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "retention_policies.*", map[string]string{
						"name":     "1day",
						"duration": "24h0m0s",
						"default":  "false",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDatabaseWithRPSUpdateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(resourceName),
					resource.TestCheckResourceAttr("influxdb_database.test", "name", rName),
					resource.TestCheckResourceAttr(resourceName, "retention_policies.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "retention_policies.*", map[string]string{
						"name":     "2days",
						"duration": "48h0m0s",
						"default":  "false",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "retention_policies.*", map[string]string{
						"name":     "12weeks",
						"duration": "2016h0m0s",
						"default":  "false",
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

func testAccDatabaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "influxdb_database" "test" {
  name = %[1]q
}
`, rName)
}

func testAccDatabaseWithRPSConfig(rName string) string {
	return fmt.Sprintf(`
resource "influxdb_database" "test" {
  name = %[1]q
  retention_policies {
   name = "1day"
    duration = "24h0m0s"
    default = "false"
  }
}
`, rName)
}

func testAccDatabaseWithRPSUpdateConfig(rName string) string {
	return fmt.Sprintf(`
resource "influxdb_database" "test" {
  name = %[1]q
  
  retention_policies {
    name     = "2days"
    duration = "48h0m0s"
  }
  
  retention_policies {
    name     = "12weeks"
    duration = "2016h0m0s"
    default  = "false"
  }
  
  retention_policies {
    name               = "1week"
    duration           = "168h0m0s"
    shardgroupduration = "2h0m0s"
  }
  }
  `, rName)
}
