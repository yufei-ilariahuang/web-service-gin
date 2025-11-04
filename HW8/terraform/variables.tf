# Region to deploy into
variable "aws_region" {
  type    = string
  default = "us-west-2"
}

# ECR & ECS settings
variable "ecr_repository_name" {
  type    = string
  default = "cs6650l2-repo"
}

variable "service_name" {
  type    = string
  default = "CS6650L2"
}

variable "container_port" {
  type    = number
  default = 8080
}

variable "ecs_count" {
  type    = number
  default = 3
}

# How long to keep logs
variable "log_retention_days" {
  type    = number
  default = 7
}

# Database settings
variable "db_name" {
  type    = string
  default = "mydb"
}

variable "db_username" {
  type    = string
  default = "admin"
}

variable "db_password" {
  type      = string
  sensitive = true
}

# DynamoDB settings
variable "dynamodb_table_name" {
  type    = string
  default = "shopping-cart"
}
