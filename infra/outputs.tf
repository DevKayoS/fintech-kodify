output "github_actions_role_arn" {
  description = "ARN da role para GitHub Actions (usar em vars.AWS_ROLE_ARN)"
  value       = aws_iam_role.github_actions.arn
}

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
