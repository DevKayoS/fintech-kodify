# Gera ZIPs placeholder para criação inicial da Lambda.
# O código real é atualizado via `aws lambda update-function-code` no CI.
data "archive_file" "placeholder" {
  type        = "zip"
  output_path = "${path.module}/placeholder.zip"

  source {
    content  = "placeholder"
    filename = "bootstrap"
  }
}

resource "aws_lambda_function" "api" {
  function_name    = "${var.project_name}-api"
  filename         = data.archive_file.placeholder.output_path
  source_code_hash = data.archive_file.placeholder.output_base64sha256
  handler          = "bootstrap"
  runtime          = "provided.al2023"
  role             = aws_iam_role.api.arn

  environment {
    variables = {
      DATABASE_URL = var.database_url
    }
  }

  # Terraform só cria/configura a função; o código é gerenciado pelo CI.
  lifecycle {
    ignore_changes = [filename, source_code_hash]
  }
}

resource "aws_lambda_function" "telegram" {
  function_name    = "${var.project_name}-telegram"
  filename         = data.archive_file.placeholder.output_path
  source_code_hash = data.archive_file.placeholder.output_base64sha256
  handler          = "bootstrap"
  runtime          = "provided.al2023"
  role             = aws_iam_role.telegram.arn

  environment {
    variables = {
      DATABASE_URL = var.database_url
    }
  }

  lifecycle {
    ignore_changes = [filename, source_code_hash]
  }
}
