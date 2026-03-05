variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "project_name" {
  description = "Project name used to name all resources"
  type        = string
  default     = "fintech-kodify"
}

variable "api_stage" {
  description = "API Gateway stage name"
  type        = string
  default     = "prod"
}

variable "database_url" {
  description = "PostgreSQL connection string (Neon)"
  type        = string
  sensitive   = true
}
