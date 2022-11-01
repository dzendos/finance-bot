-- +goose Up
-- +goose StatementBegin

CREATE TABLE currency_rate
(
    char_code TEXT UNIQUE,
    base      TEXT,
    rate      INTEGER,
    rate_date DATE
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE currency_rate;

-- +goose StatementEnd
