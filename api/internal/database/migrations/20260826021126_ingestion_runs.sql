-- +goose Up

-- TABLES
create table ingestion_runs (
    id uuid default gen_random_uuid() primary key,
    term_id integer not null,
    created_at timestamptz default now() not null
);

alter table courses add ingestion_run_id uuid references ingestion_runs(id);
alter table classes add ingestion_run_id uuid references ingestion_runs(id);

-- credits can be fractional
alter table courses drop column credits_type;

alter table courses alter column credits_min type numeric using credits_min::numeric;
alter table courses alter column credits_max type numeric using credits_max::numeric;

alter table courses add column credits_type course_credits_type generated always as (
    case
        when credits_max = credits_min
            then 'FIXED'::course_credits_type
        else 'VARIABLE'::course_credits_type
    end
) stored;

-- fix incorrect level virtual column
alter table courses drop column level;

alter table courses add column level integer generated always as (
    substring(code, '[0-9]+')::integer
) virtual;

-- fix period column
alter table class_meet_times drop constraint period_end_after_begin;

alter table class_meet_times alter column period_begin type text using period_begin::text;
alter table class_meet_times alter column period_end type text using period_end::text;

-- +goose Down

-- will fail if there is data containing non-numeric periods
alter table class_meet_times alter column period_end type integer using period_end::integer;
alter table class_meet_times alter column period_begin type integer using period_begin::integer;

alter table class_meet_times add constraint period_end_after_begin check (
    period_end >= period_begin
);

alter table courses drop column level;

alter table courses add column level integer generated always as (
    substring(code, '[0-9]+L?$')::integer
) virtual;

-- will round credits if fractional
alter table courses drop column credits_type;

alter table courses alter column credits_max type integer using credits_max::integer;
alter table courses alter column credits_min type integer using credits_min::integer;

alter table courses add column credits_type course_credits_type generated always as (
    case
        when credits_max = credits_min
            then 'FIXED'::course_credits_type
        else 'VARIABLE'::course_credits_type
    end
) stored;

alter table classes drop column ingestion_run_id;
alter table courses drop column ingestion_run_id;

drop table ingestion_runs;