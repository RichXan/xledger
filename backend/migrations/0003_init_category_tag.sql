CREATE TABLE categories (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    name TEXT NOT NULL,
    parent_id UUID,
    archived_at TIMESTAMPTZ,
    CONSTRAINT fk_categories_parent_user FOREIGN KEY (parent_id, user_id) REFERENCES categories (id, user_id)
);

CREATE UNIQUE INDEX idx_categories_id_user
    ON categories (id, user_id);

CREATE INDEX idx_categories_user_parent
    ON categories (user_id, parent_id);

ALTER TABLE transactions
    ADD COLUMN category_id UUID,
    ADD COLUMN category_name TEXT;

ALTER TABLE transactions
    ADD CONSTRAINT fk_transactions_category_user
    FOREIGN KEY (category_id, user_id) REFERENCES categories (id, user_id);

CREATE INDEX idx_transactions_user_category
    ON transactions (user_id, category_id);

CREATE TABLE tags (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    name TEXT NOT NULL
);

CREATE UNIQUE INDEX idx_tags_id_user
    ON tags (id, user_id);

CREATE UNIQUE INDEX idx_tags_user_name_lower
    ON tags (user_id, lower(name));

CREATE TABLE transaction_tags (
    transaction_id UUID NOT NULL,
    tag_id UUID NOT NULL,
    user_id UUID NOT NULL,
    PRIMARY KEY (transaction_id, tag_id, user_id),
    CONSTRAINT fk_transaction_tags_tag_user FOREIGN KEY (tag_id, user_id) REFERENCES tags (id, user_id)
);

CREATE INDEX idx_transaction_tags_user_tag
    ON transaction_tags (user_id, tag_id);
