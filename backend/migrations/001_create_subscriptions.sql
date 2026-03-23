-- +goose Up

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE subscriptions (
    id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    service_name TEXT        NOT NULL CHECK (trim(service_name) <> ''),
    price       INTEGER     NOT NULL CHECK (price > 0),
    user_id     UUID        NOT NULL,
    start_date DATE        NOT NULL,
    end_date   DATE        NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_end_date_gte_start CHECK (
        end_date IS NULL OR end_date >= start_date
    )
);

-- start_date всегда хранится как первое число месяца (например 2025-07-01) т.к. по тз нужен только месяц и год
-- end_date   тоже первое число месяца или NULL (подписка активна)

CREATE INDEX idx_subscriptions_user_id     ON subscriptions (user_id);
CREATE INDEX idx_subscriptions_service_name ON subscriptions (service_name);
CREATE INDEX idx_subscriptions_start_date  ON subscriptions (start_date);
CREATE INDEX idx_subscriptions_end_date    ON subscriptions (end_date);

-- составной индекс: фильтр по user_id + период
CREATE INDEX idx_subscriptions_user_period  ON subscriptions (user_id, start_date, end_date);

-- +goose Down

DROP TABLE IF EXISTS subscriptions;
DROP EXTENSION IF EXISTS "uuid-ossp";
