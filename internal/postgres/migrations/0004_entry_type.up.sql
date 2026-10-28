CREATE TYPE entry_type AS ENUM (
    'deposit', -- top up from outside
    'withdrawal', -- withdrawal from account
    'transfer_in', -- transfer from another account
    'transfer_out', -- transfer to another account
    'payment', -- buy from merchant
    'refund' -- money refund from merchant
);



ALTER TABLE entries
ADD COLUMN type entry_type NOT NULL;