-- +goose up

-- remove dependency on courses_pkey first
alter table classes drop constraint classes_course_id_term_id_fkey;

-- rename old course id
alter table courses rename column id to uf_id;

-- remove old composite primary key
alter table courses drop constraint courses_pkey;

-- add new uuid primary key
alter table courses add column id uuid default gen_random_uuid() primary key;

-- replace classes.course_id
alter table classes drop column course_id;
alter table classes add column course_id uuid not null references courses(id) on delete cascade;

-- restore uniqueness constraint that depended on course_id
alter table classes add constraint classes_course_id_term_id_number_key unique (course_id, term_id, number);

-- replace old uniqueness constraint
alter table courses drop constraint course_code_term_unique;

alter table courses add constraint course_unique unique (uf_id, term_id, code, name);


-- +goose down

-- remove new classes uniqueness constraint
alter table classes drop constraint classes_course_id_term_id_number_key;

-- remove dependency on courses.id
alter table classes drop constraint classes_course_id_fkey;

-- remove new uniqueness constraint
alter table courses drop constraint course_unique;

-- restore old uniqueness constraint
alter table courses add constraint course_code_term_unique unique (code, term_id);

-- replace uuid course_id
alter table classes drop column course_id;
alter table classes add column course_id integer not null;

-- remove uuid primary key
alter table courses drop constraint courses_pkey;

alter table courses drop column id;

-- restore old composite primary key
alter table courses add constraint courses_pkey primary key (uf_id, term_id);

-- restore old classes fk
alter table classes add constraint classes_course_id_term_id_fkey foreign key (course_id, term_id) references courses(uf_id, term_id) on delete cascade;

-- rename uf_id -> id
alter table courses rename column uf_id to id;