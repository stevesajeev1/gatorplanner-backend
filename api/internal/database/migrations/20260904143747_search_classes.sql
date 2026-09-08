-- +goose Up

-- TYPES
create type period as enum (
    '1',
    '2',
    '3',
    '4',
    '5',
    '6',
    '7',
    '8',
    '9',
    '10',
    '11',
    'E1',
    'E2',
    'E3'
);

alter table class_meet_times alter column period_begin type period using period_begin::period;
alter table class_meet_times alter column period_end type period using period_end::period;

alter table courses add constraint course_code_term_unique unique (code, term_id);

-- TABLES
create table buildings (
    id uuid default gen_random_uuid() primary key,
    name text not null,
    code text not null unique
);

insert into buildings (name, code) values ('WEB', 'WEB');

alter table class_meet_times drop column building;

alter table class_meet_times add column building_id uuid references buildings(id);
alter table class_meet_times add column room text;

-- +goose Down
alter table class_meet_times drop column room;
alter table class_meet_times drop column building_id;

alter table class_meet_times add column building text not null;

drop table buildings;

alter table courses drop constraint course_code_term_unique;

alter table class_meet_times alter column period_end type text using period_end::text;
alter table class_meet_times alter column period_begin type text using period_begin::text;

drop type period;
