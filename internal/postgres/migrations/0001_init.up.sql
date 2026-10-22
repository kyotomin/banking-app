CREATE TABLE "accounts" (
  "id" UUID PRIMARY KEY DEFAULT (uuidv7()),
  "owner" text NOT NULL,
  "phone_number" text UNIQUE NOT NULL,
  "secrete_code_hash" text NOT NULL,
  "balance" numeric(19, 2) NOT NULL CHECK(balance >= 0),
  "currency" char(3) NOT NULL,
  "created_at" timestamptz DEFAULT (now())
);

CREATE TABLE "entries" (
  "id" UUID PRIMARY KEY DEFAULT (uuidv7()),
  "transfer_id" UUID NOT NULL REFERENCES transfers(id),
  "account_id" UUID NOT NULL REFERENCES accounts(id),
  "amount" numeric(19, 2) NOT NULL CHECK(amount <> 0),
  "created_at" timestamptz DEFAULT (now())
);

CREATE TABLE "transfers" (
  "id" UUID PRIMARY KEY DEFAULT (uuidv7()),
  "from_account_id" UUID NOT NULL REFERENCES accounts(id),
  "to_account_id" UUID NOT NULL REFERENCES accounts(id),
  "currency" CHAR(3) NOT NULL,
  "amount" numeric(19, 2) NOT NULL,
  "created_at" timestamptz DEFAULT (now()) NOT NULL,
  "idempotency_key" text UNIQUE,
  CONSTRAINT c CHECK(from_account_id <> to_account_id)
);

CREATE INDEX ON "entries" ("account_id");

CREATE INDEX ON "transfers" ("to_account_id");

CREATE INDEX ON "transfers" ("from_account_id");
