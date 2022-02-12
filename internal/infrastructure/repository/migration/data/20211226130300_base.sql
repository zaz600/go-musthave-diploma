-- +goose Up
CREATE SCHEMA IF NOT EXISTS gophermart;
-- DROP SCHEMA gophermart CASCADE ;
-- CREATE SCHEMA gophermart;
SET SEARCH_PATH TO gophermart;

CREATE TABLE IF NOT EXISTS users
(
    id           serial primary key,
    uid          varchar ,
    login        varchar,
    password     varchar,
    created_at   TIMESTAMP
);
ALTER TABLE users ALTER COLUMN created_at SET DEFAULT now();
CREATE UNIQUE INDEX users_login_uniq_idx ON users USING btree (login);

CREATE TABLE IF NOT EXISTS sessions
(
    id           serial primary key,
    uid          varchar ,
    sid          varchar,
    created_at   TIMESTAMP
);
ALTER TABLE sessions ALTER COLUMN created_at SET DEFAULT now();
CREATE UNIQUE INDEX sessions_sid_uniq_idx ON sessions USING btree (sid);

CREATE TABLE IF NOT EXISTS orders
(
    id           serial primary key,
    uid          varchar ,
    order_id     varchar,
    uploaded_at  TIMESTAMP,
    status       varchar,
    accrual      decimal (15,2),
    retry_count  int
);
ALTER TABLE orders ALTER COLUMN uploaded_at SET DEFAULT now();
CREATE UNIQUE INDEX order_uniq_idx ON orders USING btree (order_id);

CREATE TABLE IF NOT EXISTS withdrawals
(
    id           serial primary key,
    uid          varchar ,
    order_id     varchar,
    processed_at  TIMESTAMP,
    amount      decimal (15,2)
);
ALTER TABLE withdrawals ALTER COLUMN processed_at SET DEFAULT now();
CREATE UNIQUE INDEX withdrawals_order_id_uniq_idx ON withdrawals USING btree (order_id);
