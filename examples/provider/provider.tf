
locals {
  master = jsondecode(data.aws_secretsmanager_secret_version.this.secret_string)
  host   = local.master["host"]
  port   = local.master["port"]
}

/* Secret should have the format :
{
    dbname = "metrics",
    engine =  "influxdb"
    port = 8083
    host = "myserver.com
    username = "admin"
    password = "ghdjskjdsOdfiQ!8"
  }
*/
data "aws_secretsmanager_secret" "this" {
  arn = "arn secret"
}


provider "influxdb" {
  url      = "http://${local.host}:${local.port}/"
  username = local.master["username"]
  password = local.master["password"]
}

