-- Tabela criada pelo Tern para controle de migrations.
-- Este arquivo existe apenas para o SQLC conhecer o schema.
CREATE TABLE schema_version (
    version_id bigint PRIMARY KEY,
    tstamp     timestamptz NOT NULL DEFAULT now()
);
