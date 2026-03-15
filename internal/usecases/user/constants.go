package user

const (
	// Códigos de erro PostgreSQL
	pgErrUniqueViolation = "23505"

	// Nomes das constraints de unicidade na tabela users
	pgConstraintCPF   = "cpf"
	pgConstraintEmail = "email"

	pgConstraintCPFKey   = "users_cpf_key"
	pgConstraintEmailKey = "users_email_key"
)
