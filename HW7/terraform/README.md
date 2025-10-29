# Application Load Balancer (ALB) Setup Guide

## What Was Added

Your Terraform configuration now includes a complete ALB setup following AWS best practices:

### 1. Network Architecture (modules/network/)

**Custom VPC**: `10.0.0.0/16`
- **Public Subnets** (for ALB):
  - `10.0.1.0/24` (AZ 1)
  - `10.0.2.0/24` (AZ 2)
- **Private Subnets** (for ECS):
  - `10.0.10.0/24` (AZ 1)
  - `10.0.11.0/24` (AZ 2)

**Components**:
- Internet Gateway (IGW) for public internet access
- 2 NAT Gateways (one per AZ) for private subnet internet access
- Route tables properly configured
- Security groups for ALB and ECS

### 2. Security Groups

**ALB Security Group**:
```
Inbound:  Port 80 from 0.0.0.0/0 (internet)
Outbound: All traffic
```

**ECS Security Group**:
```
Inbound:  Port 8080 from ALB only (security!)
Outbound: All traffic (for ECR pulls)
```

### 3. ALB Module (modules/alb/)

**Application Load Balancer**:
- Internet-facing
- Deployed in public subnets across 2 AZs
- HTTP listener on port 80

**Target Group**:
- Type: `ip` (required for Fargate)
- Protocol: HTTP
- Port: 8080
- Health check: `/health` endpoint
  - Interval: 30s
  - Timeout: 5s
  - Healthy threshold: 2
  - Unhealthy threshold: 3

**Traffic Flow**:
```
Internet → ALB (port 80) → Target Group → ECS Tasks (port 8080)
```

### 4. ECS Service Updates

- Tasks deployed in **private subnets**
- `assign_public_ip = false` (use NAT gateway instead)
- Integrated with ALB target group
- Auto-registration/deregistration with target group

## Directory Structure

```
terraform/
├── main.tf
├── variables.tf
├── outputs.tf
├── modules/
│   ├── network/
│   │   ├── main.tf       # VPC, subnets, IGW, NAT, SGs
│   │   ├── variables.tf
│   │   └── outputs.tf
│   ├── alb/
│   │   ├── main.tf       # ALB, target group, listener
│   │   ├── variables.tf
│   │   └── outputs.tf
│   ├── ecs/
│   │   ├── main.tf       # Cluster, service, task def
│   │   ├── variables.tf
│   │   └── outputs.tf
│   ├── ecr/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   └── outputs.tf
│   └── logging/
│       ├── main.tf
│       ├── variables.tf
│       └── outputs.tf
└── src/
    ├── main.go
    └── Dockerfile
```

## Deployment Steps

### 1. Initialize Terraform
```bash
cd terraform
terraform init
```

### 2. Plan the Deployment
```bash
terraform plan
```

**Expected Resources**: ~30 resources will be created:
- VPC, subnets, route tables, IGW, NAT gateways
- Security groups
- ALB, target group, listener
- ECS cluster, service, task definition
- ECR repository
- CloudWatch log group

### 3. Deploy
```bash
terraform apply
```

Type `yes` when prompted.

### 4. Get the ALB URL
```bash
terraform output alb_url
# Output: http://CS6650L2-alb-1234567890.us-west-2.elb.amazonaws.com
```

### 5. Test Your Service
```bash
# Get ALB URL
ALB_URL=$(terraform output -raw alb_dns_name)

# Test health endpoint
curl http://$ALB_URL/health

# Test products endpoint
curl http://$ALB_URL/products

# Test order creation
curl -X POST http://$ALB_URL/orders \
  -H "Content-Type: application/json" \
  -d '{"product_id": "1", "quantity": 2}'

# Get stats
curl http://$ALB_URL/stats
```

## How It Works

### Request Flow

1. **User** → Makes HTTP request to ALB DNS
2. **ALB** → Receives request on port 80
3. **ALB** → Routes to healthy ECS task via target group
4. **ECS Task** → Processes request on port 8080
5. **ECS Task** → Returns response
6. **ALB** → Forwards response to user

### Health Checks

- ALB continuously checks `/health` on all ECS tasks
- Unhealthy tasks are automatically removed from rotation
- New tasks are added once they pass health checks

### High Availability

- ALB deployed across 2 availability zones
- ECS tasks can scale across 2 private subnets
- If one AZ fails, traffic automatically routes to the other

## Scaling

### Manual Scaling
```bash
# Update ecs_count in variables.tf or via command line
terraform apply -var="ecs_count=3"
```

### Auto Scaling (Future Enhancement)
Add to modules/ecs/main.tf:
```hcl
resource "aws_appautoscaling_target" "ecs" {
  max_capacity       = 10
  min_capacity       = 2
  resource_id        = "service/${aws_ecs_cluster.this.name}/${aws_ecs_service.this.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

resource "aws_appautoscaling_policy" "cpu" {
  name               = "${var.service_name}-cpu-scaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.ecs.resource_id
  scalable_dimension = aws_appautoscaling_target.ecs.scalable_dimension
  service_namespace  = aws_appautoscaling_target.ecs.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
    target_value = 70.0
  }
}
```

## Monitoring

### CloudWatch Logs
```bash
# View logs
aws logs tail /ecs/CS6650L2 --follow
```

### ALB Metrics
- Target response time
- Request count
- HTTP error codes
- Healthy/unhealthy host count

### ECS Metrics
- CPU utilization
- Memory utilization
- Task count

## Cost Considerations

**Approximate monthly costs** (us-west-2):
- ALB: ~$16.20/month
- 2 NAT Gateways: ~$64.80/month ($32.40 each)
- ECS Fargate (1 task): ~$7.50/month
- Data transfer: Variable

**Total**: ~$88.50/month (excluding data transfer)

### Cost Optimization
1. **Remove second NAT Gateway** (single AZ):
   - Modify `modules/network/main.tf`
   - Use only 1 NAT gateway for both AZs
   - Saves ~$32.40/month
   - Trade-off: No HA for outbound internet

2. **Use NAT Instance** instead of NAT Gateway:
   - t4g.nano: ~$3/month
   - Saves ~$60/month
   - Trade-off: More management, lower performance

## Troubleshooting

### Tasks not starting
```bash
# Check task status
aws ecs describe-tasks \
  --cluster CS6650L2-cluster \
  --tasks $(aws ecs list-tasks --cluster CS6650L2-cluster --query 'taskArns[0]' --output text)

# Common issues:
# - Cannot pull ECR image: Check execution role permissions
# - Health check failing: Ensure /health endpoint works
# - No space in private subnets: Check subnet CIDR blocks
```

### ALB showing 503 errors
```bash
# Check target health
aws elbv2 describe-target-health \
  --target-group-arn $(terraform output -raw target_group_arn)

# Common issues:
# - No healthy targets: Check health check configuration
# - Security group blocking: Verify SG rules
# - Tasks not registered: Check ECS service configuration
```

### Cannot access ALB
```bash
# Verify ALB is active
aws elbv2 describe-load-balancers \
  --names CS6650L2-alb

# Check security group
aws ec2 describe-security-groups \
  --filters "Name=group-name,Values=CS6650L2-alb-sg"
```

## Clean Up

```bash
# Destroy all resources
terraform destroy

# Type 'yes' when prompted
```

**Note**: NAT Gateways and EIPs can take several minutes to delete.

## Next Steps

1. **Add HTTPS**: Configure ACM certificate and HTTPS listener
2. **Add Auto Scaling**: Implement CPU/memory-based scaling
3. **Add WAF**: Protect against common web exploits
4. **Add CloudFront**: CDN for static assets
5. **Add Route53**: Custom domain name
6. **Add monitoring dashboards**: CloudWatch custom dashboards

Your infrastructure is now production-ready! 🚀