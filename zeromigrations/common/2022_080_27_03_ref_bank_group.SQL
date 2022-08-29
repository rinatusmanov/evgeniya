create table billing.ref_bank_groups (
    id text primary key,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now(),
    currency_id int,
    CONSTRAINT bank_group_currency_id_fk FOREIGN KEY (currency_id) REFERENCES billing.ref_currency(id) ON DELETE CASCADE
);