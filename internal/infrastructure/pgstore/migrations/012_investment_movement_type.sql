ALTER TABLE investments
    ADD COLUMN movement_type VARCHAR(10) NOT NULL DEFAULT 'deposit';

---- create above / drop below ----

ALTER TABLE investments DROP COLUMN movement_type;
