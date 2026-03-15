package telegram

func helpMessage() string {
	return `*Kodify Bot — Comandos disponíveis*

💰 *Despesas*
/gasto <valor> <categoria> <descrição>
_Ex: /gasto 39.90 alimentacao Almoço no restaurante_

💵 *Receitas*
/receber <valor> <descrição>
_Ex: /receber 5000 Salário de março_

📈 *Investimentos*
/investimento <valor> <tipo> <descrição>
_Ex: /investimento 500 cdb Tesouro Selic_

📊 *Consultas*
/resumo — totais do mês (gastos, receitas e saldo)
/resumo\_mensal — detalhado por categoria
/extrato — histórico de transações
/categorias — lista de categorias de despesa
/tipos\_investimento — lista de tipos de investimento

🔗 *Conta*
/start — cria sua conta

❓ /ajuda — exibe esta mensagem`
}
