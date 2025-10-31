variable "service_name" {
  description = "Base name for network resources"
  type        = string
}

variable "container_port" {
  description = "Port to expose on the security group"
  type        = number
}

variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "public_subnet_cidrs" {
  description = "CIDR blocks for public subnets (ALB)"
  type        = list(string)
  default     = ["10.0.1.0/24", "10.0.2.0/24"]
}

variable "private_subnet_cidrs" {
  description = "CIDR blocks for private subnets (ECS)"
  type        = list(string)
  default     = ["10.0.10.0/24", "10.0.11.0/24"]
}