# ── GitHub Actions OIDC ───────────────────────────────────────────────────────
# Permite que o GitHub Actions assuma uma role AWS sem credenciais estáticas.

locals {
  github_repo = "DevKayoS/fintech-kodify"
}

resource "aws_iam_openid_connect_provider" "github" {
  url = "https://token.actions.githubusercontent.com"

  client_id_list = ["sts.amazonaws.com"]

  # Thumbprint da CA raiz do GitHub OIDC (bem conhecido e estável)
  thumbprint_list = ["6938fd4d98bab03faadb97b34396831e3780aea1"]
}

data "aws_iam_policy_document" "github_actions_assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRoleWithWebIdentity"]

    principals {
      type        = "Federated"
      identifiers = [aws_iam_openid_connect_provider.github.arn]
    }

    condition {
      test     = "StringEquals"
      variable = "token.actions.githubusercontent.com:aud"
      values   = ["sts.amazonaws.com"]
    }

    # Restringe ao repo (branch main ou environment production)
    condition {
      test     = "StringLike"
      variable = "token.actions.githubusercontent.com:sub"
      values = [
        "repo:${local.github_repo}:ref:refs/heads/main",
        "repo:${local.github_repo}:environment:production",
      ]
    }
  }
}

resource "aws_iam_role" "github_actions" {
  name               = "${var.project_name}-github-actions"
  assume_role_policy = data.aws_iam_policy_document.github_actions_assume.json
}

# ── Política 1: Lambda ────────────────────────────────────────────────────────

data "aws_iam_policy_document" "github_lambda" {
  statement {
    sid    = "LambdaManage"
    effect = "Allow"
    actions = [
      "lambda:CreateFunction",
      "lambda:DeleteFunction",
      "lambda:GetFunction",
      "lambda:GetFunctionConfiguration",
      "lambda:UpdateFunctionCode",
      "lambda:UpdateFunctionConfiguration",
      "lambda:ListVersionsByFunction",
      "lambda:PublishVersion",
      "lambda:AddPermission",
      "lambda:RemovePermission",
      "lambda:GetPolicy",
      "lambda:ListTags",
      "lambda:TagResource",
      "lambda:UntagResource",
    ]
    resources = [
      "arn:aws:lambda:${var.aws_region}:${data.aws_caller_identity.current.account_id}:function:${var.project_name}-*",
    ]
  }

  statement {
    sid    = "IamPassRole"
    effect = "Allow"
    actions = [
      "iam:PassRole",
      "iam:GetRole",
      "iam:CreateRole",
      "iam:DeleteRole",
      "iam:AttachRolePolicy",
      "iam:DetachRolePolicy",
      "iam:ListRolePolicies",
      "iam:ListAttachedRolePolicies",
      "iam:PutRolePolicy",
      "iam:DeleteRolePolicy",
      "iam:GetRolePolicy",
      "iam:CreateOpenIDConnectProvider",
      "iam:DeleteOpenIDConnectProvider",
      "iam:GetOpenIDConnectProvider",
      "iam:UpdateOpenIDConnectProviderThumbprint",
      "iam:TagOpenIDConnectProvider",
    ]
    resources = [
      "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.project_name}-*",
      "arn:aws:iam::${data.aws_caller_identity.current.account_id}:oidc-provider/token.actions.githubusercontent.com",
    ]
  }
}

resource "aws_iam_policy" "github_lambda" {
  name   = "${var.project_name}-github-lambda"
  policy = data.aws_iam_policy_document.github_lambda.json
}

resource "aws_iam_role_policy_attachment" "github_lambda" {
  role       = aws_iam_role.github_actions.name
  policy_arn = aws_iam_policy.github_lambda.arn
}

# ── Política 2: API Gateway ───────────────────────────────────────────────────

data "aws_iam_policy_document" "github_apigw" {
  statement {
    sid    = "ApiGatewayManage"
    effect = "Allow"
    actions = [
      "apigateway:GET",
      "apigateway:POST",
      "apigateway:PUT",
      "apigateway:PATCH",
      "apigateway:DELETE",
    ]
    resources = ["arn:aws:apigateway:${var.aws_region}::/*"]
  }
}

resource "aws_iam_policy" "github_apigw" {
  name   = "${var.project_name}-github-apigw"
  policy = data.aws_iam_policy_document.github_apigw.json
}

resource "aws_iam_role_policy_attachment" "github_apigw" {
  role       = aws_iam_role.github_actions.name
  policy_arn = aws_iam_policy.github_apigw.arn
}

# ── Política 3: Terraform state (S3) + CloudWatch Logs ───────────────────────

data "aws_iam_policy_document" "github_infra" {
  statement {
    sid    = "TerraformState"
    effect = "Allow"
    actions = [
      "s3:GetObject",
      "s3:PutObject",
      "s3:DeleteObject",
      "s3:ListBucket",
      "s3:GetBucketVersioning",
    ]
    resources = [
      "arn:aws:s3:::kodify-terraform-state",
      "arn:aws:s3:::kodify-terraform-state/*",
    ]
  }

  statement {
    sid    = "CloudWatchLogs"
    effect = "Allow"
    actions = [
      "logs:CreateLogGroup",
      "logs:DeleteLogGroup",
      "logs:DescribeLogGroups",
      "logs:ListTagsLogGroup",
      "logs:ListTagsForResource",
      "logs:PutRetentionPolicy",
      "logs:TagResource",
      "logs:UntagResource",
    ]
    resources = [
      "arn:aws:logs:${var.aws_region}:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/${var.project_name}-*",
    ]
  }
}

resource "aws_iam_policy" "github_infra" {
  name   = "${var.project_name}-github-infra"
  policy = data.aws_iam_policy_document.github_infra.json
}

resource "aws_iam_role_policy_attachment" "github_infra" {
  role       = aws_iam_role.github_actions.name
  policy_arn = aws_iam_policy.github_infra.arn
}
