
resource "aws_dynamodb_table" "shopping_carts" {
  name           = "${var.service_name}-shopping-carts"
  billing_mode   = "PAY_PER_REQUEST"  # Auto-scaling, good for testing
  hash_key       = "pk"
  range_key      = "sk"

  attribute {
    name = "pk"
    type = "S"
  }

  attribute {
    name = "sk"
    type = "S"
  }

  attribute {
    name = "customer_id"
    type = "N"
  }

  global_secondary_index {
    name               = "CustomerIdIndex"
    hash_key          = "customer_id"
    range_key         = "sk"
    projection_type   = "ALL"
    read_capacity     = 5
    write_capacity    = 5
  }

  tags = {
    Name = "${var.service_name}-shopping-carts"
  }
}

