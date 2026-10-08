CREATE TABLE auth_challenges (
    "id" UUID PRIMARY KEY DEFAULT uuidv7(),
    "phone_number" text NOT NULL,
    "code_hash" text NOT NULL,
    "expires_at" timestamptz NOT NULL,
    "attempts" int NOT NULL DEFAULT 0,
    "created_at" timestamptz NOT NULL DEFAULT now(),
);