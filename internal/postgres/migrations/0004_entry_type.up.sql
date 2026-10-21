CREATE TYPE entry_type AS ENUM (
    'deposit',
    'withdrawal',
    'transfer_in',
    'transfer_out'
);

ALTER TABLE entries
ADD COLUMN type entry_type NOT NULL;