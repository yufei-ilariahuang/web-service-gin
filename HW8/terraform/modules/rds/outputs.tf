
output "db_instance_endpoint" {
  value = aws_db_instance.rds_instance.endpoint
}

output "db_instance_name" {
  value = aws_db_instance.rds_instance.db_name
}

output "db_instance_username" {
  value = aws_db_instance.rds_instance.username
}

output "db_instance_password" {
  value     = aws_db_instance.rds_instance.password
  sensitive = true
}

output "db_security_group_id" {
    value = aws_security_group.rds_sg.id
}
