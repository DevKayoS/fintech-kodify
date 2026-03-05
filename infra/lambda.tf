resource "aws_lambda_function" "api" {
  function_name    = "${var.project_name}-api"
  filename         = "../lambda-api.zip"
  source_code_hash = filebase64sha256("../lambda-api.zip")
  handler          = "bootstrap"
  runtime          = "provided.al2023"
  role             = aws_iam_role.api.arn

  environment {
    variables = {
      DATABASE_URL = var.database_url
    }
  }
}

resource "aws_lambda_function" "telegram" {
  function_name    = "${var.project_name}-telegram"
  filename         = "../lambda-telegram.zip"
  source_code_hash = filebase64sha256("../lambda-telegram.zip")
  handler          = "bootstrap"
  runtime          = "provided.al2023"
  role             = aws_iam_role.telegram.arn

  environment {
    variables = {
      DATABASE_URL = var.database_url
    }
  }
}
