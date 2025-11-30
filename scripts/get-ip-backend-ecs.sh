#!/bin/bash
set -e

# Requirements:
# - AWS CLI installed and configured with credentials (aws configure)
# - IAM permissions required for this script:
#   * ecs:ListTasks
#   * ecs:DescribeTasks
#   * ec2:DescribeNetworkInterfaces
# - The script assumes the cluster name `inbox-cluster` and service name
#   `inbox-backend-service` are correct and reachable from the current AWS
#   account/region. Adjust names or add region flags to the `aws` calls if
#   needed (e.g. `--region us-east-1`).
# - Warning: running this requires network/API access to AWS; avoid embedding
#   long-lived credentials in scripts.

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${YELLOW}=== Obtaining backend public IP ===${NC}\n"

# Pre-check: ensure AWS CLI is available and configured
if ! command -v aws >/dev/null 2>&1; then
  echo -e "${RED}Error: AWS CLI not found. Install it and run 'aws configure' to set credentials.${NC}"
  exit 1
fi

if ! aws sts get-caller-identity --output text >/dev/null 2>&1; then
  echo -e "${RED}Error: AWS CLI cannot perform API calls. Check credentials, profile, and region. Run 'aws configure' or set AWS_PROFILE/AWS_REGION.${NC}"
  exit 1
fi

CALLER_ACCOUNT=$(aws sts get-caller-identity --query 'Account' --output text 2>/dev/null || echo "unknown")
echo -e "${GREEN}AWS caller account: ${CALLER_ACCOUNT}${NC}\n"

# 1. Get the project root directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Verify directory structure
if [ ! -d "$PROJECT_ROOT/infra" ]; then
    echo -e "${RED}Error: infra directory not found${NC}"
    exit 1
fi

#cd "$PROJECT_ROOT/infra"
PUBLIC_IP=$(aws ecs describe-tasks \
  --cluster inbox-cluster \
  --tasks $(aws ecs list-tasks --cluster inbox-cluster --service-name inbox-backend-service --query 'taskArns[0]' --output text) \
  --query 'tasks[0].attachments[0].details[?name==`networkInterfaceId`].value' \
  --output text | xargs -I {} aws ec2 describe-network-interfaces \
  --network-interface-ids {} \
  --query 'NetworkInterfaces[0].Association.PublicIp' \
  --output text)

echo -e "${GREEN}✓ The container's public IP is: ${PUBLIC_IP}\n"

TASK_ARN=$(aws ecs list-tasks --cluster inbox-cluster --service-name inbox-backend-service --query 'taskArns[0]' --output text)
echo -e "${GREEN}Task ARN: $TASK_ARN"

ENI_ID=$(aws ecs describe-tasks --cluster inbox-cluster --tasks $TASK_ARN --query 'tasks[0].attachments[0].details[?name==`networkInterfaceId`].value' --output text)
echo -e "${GREEN}Network Interface: $ENI_ID"

PUBLIC_IP=$(aws ec2 describe-network-interfaces --network-interface-ids $ENI_ID --query 'NetworkInterfaces[0].Association.PublicIp' --output text)
echo -e "${GREEN}Public IP: $PUBLIC_IP"