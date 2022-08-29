create table public.users (
    id serial primary key,
    created_at timestamp default now(),
    name varchar(255) not null
);