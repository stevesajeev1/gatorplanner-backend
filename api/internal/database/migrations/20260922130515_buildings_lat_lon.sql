-- +goose Up

alter table buildings add column latitude numeric(2,6);
alter table buildings add column longitude numeric(3,6);

-- +goose Down

alter table buildings drop column longitude;
alter table buildings drop column latitude;
