-- +goose Up
-- +goose StatementBegin

CREATE TABLE limits 
(
    tg_user_id BIGINT REFERENCES users (tg_user_id),
    month_no   INTEGER,
    user_limit INTEGER
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE limits;

-- +goose StatementEnd
