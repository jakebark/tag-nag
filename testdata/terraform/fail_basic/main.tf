resource "aws_s3_bucket" "this" {
  bucket = "fail"
  tags = {
    Owner = "test-user"
    // missing Environment
    // missing Project
  }
}
