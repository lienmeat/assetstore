resource "aws_dynamodb_table" "asset-table" {
  name           = "assets"
  billing_mode   = "PROVISIONED"
  read_capacity  = "${var.dynamo_read_capacity}"
  write_capacity = "${var.dynamo_write_capacity}"
  hash_key       = "ObjID"
  range_key      = "ObjSort"

  attribute {
    name = "ObjID"
    type = "S"
  }

  attribute {
    name = "ObjSort"
    type = "S"
  }

  tags = {
    Name        = "dynamodb-table-asset-meta"
    Environment = "${var.environment}"
  }
}