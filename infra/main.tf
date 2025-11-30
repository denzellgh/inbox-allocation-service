resource "aws_vpc" "main_vpc" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "inbox-vpc"
  }
}

resource "aws_internet_gateway" "main_internet_gateway" {
  vpc_id = aws_vpc.main_vpc.id

  tags = {
    Name = "inbox-igw"
  }
}

resource "aws_subnet" "public_snet_a" {
  vpc_id                  = aws_vpc.main_vpc.id
  cidr_block              = "10.0.1.0/24"
  availability_zone       = "us-east-1a"
  map_public_ip_on_launch = true

  tags = {
    Name = "inbox-public-subnet-1a"
  }
}

resource "aws_subnet" "public_snet_b" {
  vpc_id                  = aws_vpc.main_vpc.id
  cidr_block              = "10.0.2.0/24"
  availability_zone       = "us-east-1b"
  map_public_ip_on_launch = true

  tags = {
    Name = "inbox-public-subnet-1b"
  }
}

# Route Table
resource "aws_route_table" "public_rt" {
  vpc_id = aws_vpc.main_vpc.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main_internet_gateway.id
  }

  tags = {
    Name = "inbox-public-rt"
  }
}

resource "aws_route_table_association" "public_table_a" {
  subnet_id      = aws_subnet.public_snet_a.id
  route_table_id = aws_route_table.public_rt.id
}

resource "aws_route_table_association" "public_table_b" {
  subnet_id      = aws_subnet.public_snet_b.id
  route_table_id = aws_route_table.public_rt.id
}

#===========================================================

# Subnet Group for DB
resource "aws_db_subnet_group" "main_db_snet_gp" {
  name       = "inbox-db-subnet"
  subnet_ids = [aws_subnet.public_snet_a.id, aws_subnet.public_snet_b.id]

  tags = {
    Name = "inbox-db-subnet-group"
  }
}

# Security Group for DB
resource "aws_security_group" "database" {
  name   = "inbox-db-sg"
  vpc_id = aws_vpc.main_vpc.id

  ingress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# RDS PostgreSQL
resource "aws_db_instance" "postgres_db_instance" {

  identifier                 = "inbox-db"
  auto_minor_version_upgrade = true

  engine         = "postgres"
  engine_version = "16.11"
  instance_class = "db.t3.micro"

  allocated_storage = 20
  storage_type      = "gp2"

  db_name  = var.DB_NAME
  username = var.DB_USER
  password = var.DB_PASSWORD

  vpc_security_group_ids = [aws_security_group.database.id]
  db_subnet_group_name   = aws_db_subnet_group.main_db_snet_gp.name

  publicly_accessible = true

  backup_retention_period = 0
  skip_final_snapshot     = true

  tags = {
    Name = "inbox-database"
  }
}

resource "aws_ecr_repository" "backend_ecr" {
  name = "inbox-backend"
  force_delete = true

  depends_on = [aws_db_instance.postgres_db_instance]

  image_scanning_configuration {
    scan_on_push = false
  }

  tags = {
    Name = "inbox-backend-ecr"
  }
}

resource "null_resource" "backend_image" {
  depends_on = [aws_ecr_repository.backend_ecr]

  provisioner "local-exec" {
    command = "aws sts get-caller-identity || aws configure"
  }

  provisioner "local-exec" {
    command = "aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin ${aws_ecr_repository.backend_ecr.repository_url}"
  }

  provisioner "local-exec" {
    working_dir = "../backend"
    command     = "docker build -t ${aws_ecr_repository.backend_ecr.repository_url}:latest ."
  }

  provisioner "local-exec" {
    command = "docker push ${aws_ecr_repository.backend_ecr.repository_url}:latest"
  }

  triggers = {
    ecr_id = aws_ecr_repository.backend_ecr.id
  }
}

resource "aws_security_group" "ecs_tasks_sg" {
  name   = "inbox-backend-sg"
  description = "Security group for ECS tasks"
  vpc_id = aws_vpc.main_vpc.id

  # HTTP
  ingress {
    description = "HTTP from anywhere"
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Salida
  egress {
    description = "Allow all outbound traffic"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "inbox-ecs-tasks-sg"
  }
}

# ============================================
# 2. IAM ROLES - Permisos para ECS
# ============================================

# Role para que ECS pueda ejecutar tasks
resource "aws_iam_role" "ecs_task_execution_role" {
  name = "inbox-ecs-task-execution-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name = "inbox-ecs-task-execution-role"
  }
}

# Política para que ECS pueda hacer pull de ECR y escribir logs
resource "aws_iam_role_policy_attachment" "ecs_task_execution_role_policy" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# Role para la aplicación (si tu app necesita acceder a otros servicios AWS)
resource "aws_iam_role" "ecs_task_role" {
  name = "inbox-ecs-task-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name = "inbox-ecs-task-role"
  }
}

# ============================================
# 3. CLOUDWATCH LOGS - Para ver los logs del contenedor
# ============================================
resource "aws_cloudwatch_log_group" "ecs_logs" {
  name              = "/ecs/inbox-backend"
  retention_in_days = 7 # Free tier permite hasta 5GB/mes

  tags = {
    Name = "inbox-backend-logs"
  }
}

# ============================================
# 4. ECS CLUSTER - Contenedor lógico
# ============================================
resource "aws_ecs_cluster" "main" {
  name = "inbox-cluster"

  tags = {
    Name = "inbox-cluster"
  }
}

# ============================================
# 5. TASK DEFINITION - Define cómo ejecutar el contenedor
# ============================================
resource "aws_ecs_task_definition" "backend" {
  family                   = "inbox-backend"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn            = aws_iam_role.ecs_task_role.arn

  depends_on = [aws_db_instance.postgres_db_instance, aws_ecr_repository.backend_ecr]

  container_definitions = jsonencode([
    {
      name      = "backend"
      image     = "${aws_ecr_repository.backend_ecr.repository_url}:latest"
      essential = true

      portMappings = [
        {
          containerPort = 8080
          hostPort      = 8080
          protocol      = "tcp"
        }
      ]

      # AQUÍ DEFINES TUS VARIABLES DE ENTORNO (similar a Azure)
      environment = [
        {
          name  = "DB_HOST"
          value = aws_db_instance.postgres_db_instance.address
        },
        {
          name  = "DB_PORT"
          value = "5432"
        },
        {
          name  = "DB_USER"
          value = var.DB_USER
        },
        {
          name  = "DB_PASSWORD"
          value = var.DB_PASSWORD
        },
        {
          name  = "DB_NAME"
          value = "allocation_db"
        },
        {
          name  = "DB_SSL_MODE"
          value = "require"
        },
        {
          name  = "DB_MAX_CONNS"
          value = var.DB_MAX_CONNS
        },
        {
          name  = "DB_MIN_CONNS"
          value = var.DB_MIN_CONNS
        },
        {
          name  = "LOG_LEVEL"
          value = "debug"
        },
        {
          name  = "LOG_FORMAT"
          value = "json"
        },
        {
          name  = "READ_TIMEOUT"
          value = "15s"
        },
        {
          name  = "WRITE_TIMEOUT"
          value = "15s"
        },
        {
          name  = "IDLE_TIMEOUT"
          value = "60s"
        },
        {
          name  = "SHUTDOWN_TIMEOUT"
          value = "30s"
        },
        {
          name  = "GRACE_PERIOD_INTERVAL"
          value = "120s"
        },
        {
          name  = "GRACE_PERIOD_BATCH_SIZE"
          value = "100"
        },
        {
          name  = "IDEMPOTENCY_TTL"
          value = "24h"
        },
        {
          name  = "IDEMPOTENCY_CLEANUP_INTERVAL"
          value = "1h"
        }
      ]

      # Configuración de logs
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.ecs_logs.name
          "awslogs-region"        = "us-east-1"
          "awslogs-stream-prefix" = "backend"
        }
      }
    }
  ])

  tags = {
    Name = "inbox-backend-task"
  }
}

# ============================================
# ECS SERVICE - Keep container running
# ============================================
resource "aws_ecs_service" "backend" {
  name            = "inbox-backend-service"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.backend.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = [aws_subnet.public_snet_a.id]
    security_groups  = [aws_security_group.ecs_tasks_sg.id]
    assign_public_ip = true
  }

  tags = {
    Name = "inbox-backend-service"
  }
}