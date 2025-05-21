provider "aws" {
  region = "us-east-1"
  default_tags {
    tags = {
      Source  = "my-repo"
      Project = "112233"
    }
  }
}

resource "aws_s3_bucket" "this" {
  bucket = "test-bucket"
  tags = {
    Owner       = "jakebark"
    Environment = "dev"
  }
}
