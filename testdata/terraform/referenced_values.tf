resource "aws_s3_bucket" "this" {
  bucket = "test-bucket"
  tags = {
    Owner       = var.owner
    Environment = local.environment
  }
}

variable "owner" {
  type    = string
  default = "jakebark"
}

locals {
  environment = "dev"
}
