-- +goose Up
-- +goose StatementBegin
CREATE TABLE orders (
    id   varchar(19) constraint order_pk PRIMARY KEY,
    data JSONB,
    created_at timestamp
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE orders;
-- +goose StatementEnd