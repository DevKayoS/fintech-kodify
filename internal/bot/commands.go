package bot

func helpMessage() string {
	return `*Kodify Bot — Comandos disponíveis*

💰 *Despesas*
/gasto <valor> <categoria> <descrição>
_Ex: /gasto 39.90 alimentacao Almoço no restaurante_

📈 *Investimentos*
/investimento <valor> <tipo> <descrição>
_Ex: /investimento 500 cdb Tesouro Selic_

📊 *Consultas*
/resumo — resumo financeiro do mês atual
/extrato — histórico de transações
/categorias — lista de categorias de despesa
/tipos\_investimento — lista de tipos de investimento

🔗 *Conta*
/start <token> — vincula este chat à sua conta

❓ /ajuda — exibe esta mensagem`
}
