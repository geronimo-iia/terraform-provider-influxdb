package influxdb

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/influxdata/influxdb/client"
)

func TestAccInfluxDBUser_admin(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resourceName := "influxdb_user.test"
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_admin(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "password"),
					resource.TestCheckResourceAttr(resourceName, "admin", "true"),
					resource.TestCheckResourceAttr(resourceName, "grant.#", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				Config: testAccUserConfig_revoke(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName),
					testAccCheckUserNoAdmin(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "password"),
					resource.TestCheckResourceAttr(resourceName, "admin", "false"),
					resource.TestCheckResourceAttr(resourceName, "grant.#", "0"),
				),
			},
		},
	})
}

func TestAccInfluxDBUser_grant(t *testing.T) {
	resourceName := "influxdb_user.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_grant(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "password"),
					resource.TestCheckResourceAttr(resourceName, "admin", "false"),
					resource.TestCheckResourceAttr(resourceName, "grant.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "grant.*", map[string]string{
						"database":  rName,
						"privilege": "READ",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				Config: testAccUserConfig_grantUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "password"),
					resource.TestCheckResourceAttr(resourceName, "admin", "false"),
					resource.TestCheckResourceAttr(resourceName, "grant.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "grant.*", map[string]string{
						"database":  rName,
						"privilege": "ALL",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "grant.*", map[string]string{
						"database":  fmt.Sprintf("%s-2", rName),
						"privilege": "WRITE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "grant.*", map[string]string{
						"database":  fmt.Sprintf("%s-3", rName),
						"privilege": "READ",
					}),
				),
			},
		},
	})
}

func TestAccInfluxDBUser_revoke(t *testing.T) {
	resourceName := "influxdb_user.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_grant(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "password"),
					resource.TestCheckResourceAttr(resourceName, "admin", "false"),
					resource.TestCheckResourceAttr(resourceName, "grant.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "grant.*", map[string]string{
						"database":  rName,
						"privilege": "READ",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				Config: testAccUserConfig_revoke(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "password"),
					resource.TestCheckResourceAttr(resourceName, "admin", "false"),
					resource.TestCheckResourceAttr(resourceName, "grant.#", "0"),
				),
			},
		},
	})
}

func testAccCheckUserExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No user id set")
		}

		conn := testAccProvider.Meta().(*client.Client)

		query := client.Query{
			Command: "SHOW USERS",
		}

		resp, err := conn.Query(query)
		if err != nil {
			return err
		}

		if resp.Err != nil {
			return resp.Err
		}

		for _, result := range resp.Results[0].Series[0].Values {
			if result[0] == rs.Primary.ID {
				return nil
			}
		}

		return fmt.Errorf("User %q does not exist", rs.Primary.ID)
	}
}

func testAccCheckUserNoAdmin(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No user id set")
		}

		conn := testAccProvider.Meta().(*client.Client)

		query := client.Query{
			Command: "SHOW USERS",
		}

		resp, err := conn.Query(query)
		if err != nil {
			return err
		}

		if resp.Err != nil {
			return resp.Err
		}

		for _, result := range resp.Results[0].Series[0].Values {
			if result[0] == rs.Primary.ID {
				if result[1].(bool) == true {
					return fmt.Errorf("User %q is admin", rs.Primary.ID)
				}

				return nil
			}
		}

		return fmt.Errorf("User %q does not exist", rs.Primary.ID)
	}
}

func testAccUserConfig_admin(rName string) string {
	return fmt.Sprintf(`
resource "influxdb_user" "test" {
  name     = %[1]q
  password = %[1]q
  admin    = true
}
`, rName)
}

func testAccUserConfig_grant(rName string) string {
	return fmt.Sprintf(`
resource "influxdb_database" "test" {
  name = %[1]q
}

resource "influxdb_user" "test" {
  name     = %[1]q
  password = %[1]q
  
  grant {
    database = influxdb_database.test.name
    privilege = "READ"
  }
}
`, rName)
}

func testAccUserConfig_revoke(rName string) string {
	return fmt.Sprintf(`	
resource "influxdb_database" "test" {
  name = %[1]q
}

resource "influxdb_user" "test" {
  name     = influxdb_database.test.name
  password = %[1]q
  admin    = false
}
`, rName)
}

func testAccUserConfig_grantUpdate(rName string) string {
	return fmt.Sprintf(`		
resource "influxdb_database" "test" {
    name = %[1]q
}

resource "influxdb_database" "test2" {
    name = "%[1]s-2"
}

resource "influxdb_database" "test3" {
    name = "%[1]s-3"
}

resource "influxdb_user" "test" {
  name     = %[1]q
  password = %[1]q

  grant {
    database = influxdb_database.test.name
    privilege = "ALL"
  }
  
  grant {
    database = influxdb_database.test2.name
    privilege = "WRITE"
  }
  
  grant {
    database = influxdb_database.test3.name
    privilege = "READ"
  }
}
`, rName)
}
