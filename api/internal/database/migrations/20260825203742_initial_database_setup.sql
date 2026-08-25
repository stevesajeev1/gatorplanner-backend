-- +goose Up

-- TYPES
create type course_credits_type as enum (
    'FIXED',
    'VARIABLE'
);

create type gen_ed as enum (
    'Biological Science',
    'Physical Science',
    'Social Science',
    'Mathematics',
    'Composition',
    'Humanities',
    'International'
);

create type quest as enum (
    'Quest 1',
    'Quest 2',
    'Quest 3',
    'Quest 4'
);

create type class_meet_type as enum (
    'Primarily Classroom',
    'Hybrid',
    'Online (80-99%)',
    'Online (100%)'
);

-- TABLES
create table departments (
    id uuid default gen_random_uuid() primary key,
    code integer not null unique,
    name text not null
);

create table courses (
    id integer not null,
    term_id integer not null,
    code text not null,

    primary key (id, term_id),

    code_prefix text generated always as (
        substring(code from '^[A-Za-z]+')
    ) virtual,
    level integer generated always as (
        substring(code from '[0-9]+L?$')::integer
    ) virtual,
    is_lab boolean generated always as (
        code ~ 'L$'
    ) virtual,

    name text not null,
    description text not null,
    syllabus text not null,
    prerequisites text not null,

    credits_min integer not null,
    credits_max integer not null,
    credits_type course_credits_type generated always as (
        case
            when credits_max = credits_min
                then 'FIXED'::course_credits_type
            else 'VARIABLE'::course_credits_type
        end
    ) stored,

    constraint credits_non_negative
        check (
            credits_min >= 0
        ),
    constraint credits_max_greater_than_min
        check (
            credits_max >= credits_min
        ),

    department_id uuid not null references departments(id),

    words integer not null,
    gen_eds gen_ed[] not null,
    quest quest,
    is_ai boolean not null,
    is_honors boolean not null
);

create table classes (
    id uuid default gen_random_uuid() primary key,
    course_id integer not null,
    term_id integer not null,
    number integer not null,

    foreign key (course_id, term_id)
        references courses(id, term_id)
        on delete cascade,
    unique (course_id, term_id, number),

    note text,

    meet_type class_meet_type not null
);

create table class_meet_times (
    id uuid default gen_random_uuid() primary key,
    class_id uuid not null references classes(id) on delete cascade,
    number integer not null,

    unique (class_id, number),

    days char(1)[] not null,
    time_begin time not null,
    time_end time not null,
    period_begin integer not null,
    period_end integer not null,

    constraint valid_days
        check (days <@ ARRAY['M','T','W','R','F']::char(1)[]),
    constraint time_end_after_begin
        check (
            time_end >= time_begin
        ),
    constraint period_end_after_begin
        check (
            period_end >= period_begin
        ),

    building text not null
);

create table instructors (
    id uuid default gen_random_uuid() primary key,
    name text not null,

    rmp_id integer unique,
    rating numeric(2,1),
    difficulty numeric(2,1),

    constraint rating_range
        check (rating between 0 and 5),
    constraint difficulty_range
        check (difficulty between 0 and 5)
);

create table class_instructors (
    class_id uuid not null references classes(id) on delete cascade,
    instructor_id uuid not null references instructors(id) on delete cascade,

    primary key (class_id, instructor_id)
);

-- +goose Down
drop table class_instructors;
drop table instructors;
drop table class_meet_times;
drop table classes;
drop table courses;
drop table departments;

drop type class_meet_type;
drop type quest;
drop type gen_ed;
drop type course_credits_type;