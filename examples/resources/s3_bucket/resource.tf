# Copyright (c) HashiCorp, Inc.

resource "s3_bucket" "example" {
  name = "example-12345"
  tags = "example"
}