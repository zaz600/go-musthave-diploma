-- +goose Up
SET SEARCH_PATH TO gophermart;

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
