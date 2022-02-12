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
CREATE UNIQUE INDEX login_uniq_idx ON users USING btree (login);

