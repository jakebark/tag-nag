resource "aws_s3_bucket" "foo" {
  bucket = "test-bucket"
}

resource "aws_s3_bucket" "bar" {
  bucket = "test-bucket"
  tags = {
    Project = "112233"
  }
}

resource "aws_s3_bucket" "baz" {
  bucket   = "test-bucket"
  provider = aws.west
}
