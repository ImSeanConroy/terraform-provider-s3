# Copyright (c) HashiCorp, Inc.

resource "customs3_bucket" "example" {
  name = "example-12345"
  tags = "example"
}