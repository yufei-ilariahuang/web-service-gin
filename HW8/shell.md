```bash
cd terraform
terraform init
terraform plan  # Review the changes
terraform apply # Deploy infrastructure

# 1. Set the correct values from your terraform output
REPO_URL="637423451078.dkr.ecr.us-west-2.amazonaws.com/ecr_service"
CLUSTER_NAME="CS6650L2-cluster"
SERVICE_NAME="CS6650L2"

# 2. Build the image (with platform specification for M1/M2 Macs)
cd ../src
docker build --platform linux/amd64 -t shopping-cart .

# 3. Login to ECR
aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin $REPO_URL

# 4. Tag and push
docker tag shopping-cart:latest $REPO_URL:latest
docker push $REPO_URL:latest

# 5. Update ECS service with the exact names
aws ecs update-service \
  --cluster "CS6650L2-cluster" \
  --service "CS6650L2" \
  --force-new-deployment \
  --region us-west-2
```