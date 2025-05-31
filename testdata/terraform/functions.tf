resource "aws_s3_bucket" "this" {
  bucket = "test-bucket"
  tags = {
    Owner       = "jakebark"
    Environment = lower("Dev")
  }
}
