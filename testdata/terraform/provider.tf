provider "aws" {
  region = "us-east-1"
  default_tags {
    tags = {
      Project = "112233"
      Source  = "my-repo"
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
