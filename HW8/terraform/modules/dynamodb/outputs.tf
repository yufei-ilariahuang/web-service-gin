
output "table_name" {
  value = aws_dynamodb_table.shopping_carts.name
  description = "Name of the created DynamoDB table"
}
