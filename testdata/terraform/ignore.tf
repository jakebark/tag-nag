resource "aws_s3_bucket" "this" {
  #tag-nag ignore
  bucket = "test-bucket"
  tags = {
    Owner       = "jakebark"
    Environment = "dev"
  }
}
