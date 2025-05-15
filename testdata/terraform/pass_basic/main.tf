resource "aws_s3_bucket" "this" {
  bucket = "pass"
  tags = {
    Owner       = "test-user"
    Environment = "dev"
    Project     = "tag-nag"
  }
}
