#!/usr/bin/env bash
# Roda UMA VEZ para deletar os recursos criados manualmente na AWS.
# Depois disso o Terraform assume o controle.
#
# Uso: AWS_REGION=us-east-1 bash scripts/destroy-aws-resources.sh

set -euo pipefail

AWS_REGION="${AWS_REGION:-us-east-1}"

echo "=== Limpando recursos AWS criados manualmente ==="
echo "Região: $AWS_REGION"
echo ""

# ── Lambda functions ──────────────────────────────────────────────────────────
for fn in "fintech-kodify-api" "fintech-kodify-telegram"; do
  echo "→ Deletando Lambda: $fn"
  aws lambda delete-function --function-name "$fn" --region "$AWS_REGION" \
    2>/dev/null && echo "  ✓ Deletado" || echo "  → Não encontrado, pulando."
done

# ── API Gateway ───────────────────────────────────────────────────────────────
echo ""
echo "→ Procurando API Gateway 'api-kodify'..."
API_ID=$(aws apigateway get-rest-apis \
  --region "$AWS_REGION" \
  --query "items[?name=='api-kodify'].id | [0]" \
  --output text 2>/dev/null || echo "None")

if [ "$API_ID" != "None" ] && [ -n "$API_ID" ]; then
  echo "  → Deletando API Gateway: $API_ID"
  aws apigateway delete-rest-api --rest-api-id "$API_ID" --region "$AWS_REGION"
  echo "  ✓ Deletado"
else
  echo "  → Não encontrado, pulando."
fi

# ── IAM roles criadas manualmente (lambda-execution-role) ────────────────────
echo ""
echo "→ Deletando IAM role: lambda-execution-role"
aws iam detach-role-policy \
  --role-name lambda-execution-role \
  --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole \
  2>/dev/null || true
aws iam delete-role --role-name lambda-execution-role \
  2>/dev/null && echo "  ✓ Deletado" || echo "  → Não encontrada, pulando."

echo ""
echo "✓ Limpeza concluída!"
echo ""
echo "Próximos passos:"
echo "  1. Crie o bucket S3 para o state (uma vez):"
echo "     aws s3 mb s3://kodify-terraform-state --region $AWS_REGION"
echo "  2. Rode o primeiro deploy:"
echo "     cd infra && terraform init && terraform apply"
echo "  Ou faça um push para a branch main e o CI cuida do resto."
