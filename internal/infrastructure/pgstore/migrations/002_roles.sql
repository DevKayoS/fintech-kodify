-- Write your migrate up statements here

CREATE TABLE roles (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE permissions (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE role_permissions (
    role_id       BIGINT REFERENCES roles(id) ON DELETE CASCADE,
    permission_id BIGINT REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

ALTER TABLE users ADD COLUMN role_id BIGINT REFERENCES roles(id);

INSERT INTO roles (name, description) VALUES
    ('admin',     'Administrator with full access'),
    ('user',      'Regular user with limited access'),
    ('moderator', 'Moderator with some admin capabilities');

INSERT INTO permissions (name, description) VALUES
    ('read:expenses',     'Can view expenses'),
    ('write:expenses',    'Can create expenses'),
    ('delete:expenses',   'Can delete expenses'),
    ('read:investments',  'Can view investments'),
    ('write:investments', 'Can create investments'),
    ('delete:investments','Can delete investments'),
    ('read:users',        'Can view users'),
    ('write:users',       'Can create users'),
    ('delete:users',      'Can delete users'),
    ('manage:all',        'Full system access');

-- Admin tem tudo
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p WHERE r.name = 'admin';

-- User tem apenas leitura e escrita de gastos e investimentos próprios
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'user'
  AND p.name IN (
      'read:expenses', 'write:expenses', 'delete:expenses',
      'read:investments', 'write:investments', 'delete:investments'
  );

-- Moderator tem permissões ampliadas
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'moderator'
  AND p.name IN (
      'read:expenses', 'write:expenses', 'delete:expenses',
      'read:investments', 'write:investments', 'delete:investments',
      'read:users', 'write:users'
  );

---- create above / drop below ----

ALTER TABLE users DROP COLUMN IF EXISTS role_id;

DELETE FROM role_permissions;
DELETE FROM permissions;
DELETE FROM roles;

DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
