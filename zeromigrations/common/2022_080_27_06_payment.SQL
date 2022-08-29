create table billing.payments
(
    id              serial primary key,
    created_at      timestamp not null default now(),
    rollback_at      timestamp default null,
    user_id         integer   not null references public.users (id),
    currency_id     integer   not null references billing.ref_currency (id),
    withdraw        integer   not null default 0,
    deposit         integer   not null default 0,
    transaction_ref text      not null default ''
);