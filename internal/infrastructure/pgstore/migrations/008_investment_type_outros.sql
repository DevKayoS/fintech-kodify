-- Write your migrate up statements here

INSERT INTO investment_types (slug, name) VALUES
    ('outros', 'Outros');

---- create above / drop below ----

DELETE FROM investment_types WHERE slug = 'outros';
