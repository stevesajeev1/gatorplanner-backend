-- +goose Up

-- TABLES
alter table instructors add take_again numeric(4,1);

alter table instructors add constraint take_again_range check (take_again between 0 and 100);

alter table instructors add created_at timestamptz default now() not null;

-- +goose Down
alter table instructors drop column created_at;

alter table instructors drop constraint take_again_range;

alter table instructors drop column take_again;
