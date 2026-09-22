-- +goose Up

alter table buildings add column latitude numeric(8,6);
alter table buildings add column longitude numeric(9,6);

alter table buildings add constraint latitude_range check (latitude between -90 and 90);
alter table buildings add constraint longitude_range check (longitude between -180 and 180);

-- +goose Down

alter table buildings drop constraint longitude_range;
alter table buildings drop constraint latitude_range;

alter table buildings drop column longitude;
alter table buildings drop column latitude;
