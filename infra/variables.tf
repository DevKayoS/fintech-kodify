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

variable "telegram_bot_token" {
  description = "Token do bot Telegram (obtido via @BotFather)"
  type        = string
  sensitive   = true
}

variable "telegram_webhook_secret" {
  description = "Segredo para validar webhooks do Telegram (X-Telegram-Bot-Api-Secret-Token)"
  type        = string
  sensitive   = true
}

variable "redis_url" {
  description = "URL de conexão do Redis (Upstash) para rate limiting — formato: rediss://:password@host:port"
  type        = string
  sensitive   = true
}
