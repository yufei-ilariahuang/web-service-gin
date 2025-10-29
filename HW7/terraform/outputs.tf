output "ecs_cluster_name" {
  description = "Name of the created ECS cluster"
  value       = module.ecs.cluster_name
}

output "ecs_service_name" {
  description = "Name of the running ECS service"
  value       = module.ecs.service_name
}

output "ecr_repository_url" {
  description = "ECR repository URL"
  value       = module.ecr.repository_url
}

output "task_public_ips" {
  description = "Public IPs of ECS tasks (check ECS console)"
  value       = "Check ECS Console > Tasks for public IPs"
}

output "alb_dns_name" {
  description = "ALB DNS name to access the service"
  value       = module.alb.alb_dns_name
}

output "alb_url" {
  description = "Full URL to access the service"
  value       = "http://${module.alb.alb_dns_name}"
}