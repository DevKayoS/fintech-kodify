-- Write your migrate up statements here

CREATE TABLE expense_categories (
    id          SERIAL PRIMARY KEY,
    slug        VARCHAR(50) UNIQUE NOT NULL,
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

INSERT INTO expense_categories (slug, name) VALUES
    ('alimentacao', 'Alimentação'),
    ('transporte',  'Transporte'),
    ('saude',       'Saúde'),
    ('lazer',       'Lazer'),
    ('moradia',     'Moradia'),
    ('educacao',    'Educação'),
    ('outros',      'Outros');

---- create above / drop below ----

DROP TABLE IF EXISTS expense_categories;
