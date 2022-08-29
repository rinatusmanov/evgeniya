insert into billing.ref_bank_groups
    (id, created_at, updated_at, currency_id)
select ('BANK_' || code) as id, now() as created_at, now() as updated_at, id as currency_id
from billing.ref_currency rc;


