-- +goose Up
-- +goose StatementBegin

CREATE TABLE users
(
    tg_user_id       BIGINT UNIQUE PRIMARY KEY,
    expense_id       INTEGER,
    current_state    TEXT,
    current_currency TEXT,

    month_no   INTEGER,
    user_limit INTEGER
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE users;

-- +goose StatementEnd
