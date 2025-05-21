resource "aws_s3_bucket" "this" {
  bucket = "test-bucket"
  tags = {
    Owner       = var.owner
    Environment = local.environment
    Project     = "${local.project}"
    Source      = "${local.source}"
  }
}

variable "owner" {
  type    = string
  default = "jakebark"
}

variable "source" {
  type    = string
  default = "my-repo"
}

locals {
  environment = "dev"
  project     = "112233"
  source      = var.source
}
