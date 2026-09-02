-- +goose Up

-- TYPES
create type meet_day_type as enum (
    'M',
    'T',
    'W',
    'R',
    'F',
    'S',
    'U'
);

alter table class_meet_times drop constraint valid_days;

alter table class_meet_times alter column days type meet_day_type[] using days::meet_day_type[];

-- +goose Down
alter table class_meet_times alter column days type char(1)[] using days::char(1)[];

alter table class_meet_times add constraint valid_days check (days <@ ARRAY['M','T','W','R','F','S','U']::char(1)[]);

drop type meet_day_type;
