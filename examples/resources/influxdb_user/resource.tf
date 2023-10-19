resource "influxdb_database" "green" {
    name = "terraform-green"
}

resource "influxdb_user" "paul" {
    name = "paul"
    password = "super-secret" # store that in a secret !

    grant {
      database = influxdb_database.green.name
      privilege = "write"
    }
}
