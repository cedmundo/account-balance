locals {
  binary_name = "bootstrap"
  binary_path = "${path.module}/../support/bin/bootstrap"
  src_path = "${path.module}/../lambda/hello-world"
  archive_path = "${path.module}/../support/proc-txns-csv-lambda.zip"
}