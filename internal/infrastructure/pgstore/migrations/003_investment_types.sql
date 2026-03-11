-- Write your migrate up statements here

CREATE TABLE investment_types (
    id          SERIAL PRIMARY KEY,
    slug        VARCHAR(50) UNIQUE NOT NULL,
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

INSERT INTO investment_types (slug, name) VALUES
    ('cdb',      'CDB'),
    ('cotas',    'Fundos de Investimento'),
    ('acoes',    'Ações'),
    ('tesouro',  'Tesouro Direto'),
    ('cripto',   'Criptomoedas'),
    ('poupanca', 'Poupança');

---- create above / drop below ----

DROP TABLE IF EXISTS investment_types;
