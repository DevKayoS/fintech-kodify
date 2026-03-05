output "api_url" {
  description = "URL base da API Gateway"
  value       = aws_api_gateway_stage.api.invoke_url
}

output "api_lambda_arn" {
  value = aws_lambda_function.api.arn
}

output "telegram_lambda_arn" {
  value = aws_lambda_function.telegram.arn
}
