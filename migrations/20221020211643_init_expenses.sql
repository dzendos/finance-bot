-- +goose Up
-- +goose StatementBegin

CREATE TABLE expenses
(
    tg_user_id  BIGINT REFERENCES users (tg_user_id),
    expense_id  INTEGER,
    expense_sum INTEGER,
    category    TEXT,
    created_at  DATE
);

-- Используем btree - id как и дату удобно сравнивать и сортировать.
CREATE INDEX expenses_user_ts_idx on expenses(tg_user_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX expenses_user_ts_idx;

DROP TABLE expenses;

-- +goose StatementEnd
