-- +goose Up

alter table courses alter column code_prefix set not null;
alter table courses alter column level set not null;
alter table courses alter column is_lab set not null;

-- +goose Down

alter table courses alter column is_lab drop not null;
alter table courses alter column level drop not null;
alter table courses alter column code_prefix drop not null;