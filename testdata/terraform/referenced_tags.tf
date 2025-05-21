resource "aws_s3_bucket" "this" {
  bucket = "test-bucket"
  tags   = var.tags
}

variable "tags" {
  type = map(string)
  default = {
    Owner       = "jakebark"
    Environment = "dev"
  }
}
