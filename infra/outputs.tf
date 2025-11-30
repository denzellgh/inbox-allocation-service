output "database_endpoint" {
  description = "Database endpoint"
  value       = aws_db_instance.postgres_db_instance.endpoint
}

output "database_address" {
  description = "Database enlace"
  value       = aws_db_instance.postgres_db_instance.address
}

output "vpc_id" {
  value = aws_vpc.main_vpc.id
}

output "publica_subnet_id" {
  value = aws_subnet.public_snet_a.id
}

output "publicb_subnet_id" {
  value = aws_subnet.public_snet_b.id
}

output "ecr_repository_url" {
  description = "URL ECR Repository"
  value       = aws_ecr_repository.backend_ecr.repository_url
}

output "name_ecr_repository_url" {
  description = "ECR Repository name"
  value       = aws_ecr_repository.backend_ecr.name
}

output "id_ecr_repository_url" {
  description = "ECR Repository ID"
  value       = aws_ecr_repository.backend_ecr.id
}

output "get_ip_backend" {
  value = "Run the script `scripts/get-ip-backend-ecs.sh` from the repository root to obtain the ECS task public IP: `bash scripts/get-ip-backend-ecs.sh`"
  description = "Instruction to run helper script that retrieves the ECS public IP"
}

output "view_logs_command" {
  value = "aws logs tail /ecs/kangaroo-backend --follow"
  description = "Command to view container logs in real-time"
}