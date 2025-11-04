variable "service_name" {
  type        = string
  description = "Base name for ECS resources"
}

variable "image" {
  type        = string
  description = "ECR image URI (with tag)"
}

variable "container_port" {
  type        = number
  description = "Port your app listens on"
}

variable "subnet_ids" {
  type        = list(string)
  description = "Subnets for FARGATE tasks"
}

variable "security_group_ids" {
  type        = list(string)
  description = "SGs for FARGATE tasks"
}

variable "execution_role_arn" {
  type        = string
  description = "ECS Task Execution Role ARN"
}

variable "db_backend" {
  type        = string
  description = "Database backend to use (mysql or dynamodb)"
  default     = "mysql"
}

variable "task_role_arn" {
  type        = string
  description = "IAM Role ARN for app permissions"
}

variable "log_group_name" {
  type        = string
  description = "CloudWatch log group name"
}

variable "ecs_count" {
  type        = number
  default     = 1
  description = "Desired Fargate task count"
}

variable "region" {
  type        = string
  description = "AWS region (for awslogs driver)"
}

variable "cpu" {
  type        = string
  default     = "256"
  description = "vCPU units"
}

variable "memory" {
  type        = string
  default     = "512"
  description = "Memory (MiB)"
}

variable "target_group_arn" {

  type        = string

  description = "ALB Target Group ARN"

}



variable "db_endpoint" {

  type        = string

  description = "RDS instance endpoint"

}



variable "db_name" {

  type        = string

  description = "RDS database name"

}



variable "db_user" {

  type        = string

  description = "RDS database user"

}



variable "db_password" {



  type        = string



  description = "RDS database password"



  sensitive   = true



}







variable "dynamodb_table_name" {



  type        = string



  description = "DynamoDB table name"



}




