variable "aws_region" {
  default = "us-east-1"
}

variable "DB_USER" {
  type = string
}

variable "DB_PASSWORD" {
  type = string
}

variable "DB_NAME" {
  type = string
}

variable "DB_MAX_CONNS" {
  type = string
}

variable "DB_MIN_CONNS" {
  type = string
}