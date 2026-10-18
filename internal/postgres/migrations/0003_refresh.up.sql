CREATE TABLE refresh_tokens (
    id UUID PRIAMRY KEY DEFAULT uuidv7(),
    account_id UUID NOT NULL REFERENCES accounts(id),
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at TIMESTAMPTZ
);