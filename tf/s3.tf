resource "aws_s3_bucket" "b" {
  bucket = "${var.unique_s3_bucket_prefix}.assets"
  acl    = "private"
}