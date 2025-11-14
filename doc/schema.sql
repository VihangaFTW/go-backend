-- SQL dump generated using DBML (dbml.dbdiagram.io)
-- Database: PostgreSQL
-- Generated at: 2025-11-14T07:08:34.455Z

CREATE TABLE "users" (
  "username" varchar PRIMARY KEY,
  "hashed_password" varchar NOT NULL,
  "full_name" varchar NOT NULL,
  "email" varchar UNIQUE NOT NULL,
  "is_email_verified" bool NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "password_changed_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "verify_emails" (
  "id" bigserial PRIMARY KEY,
  "username" varchar NOT NULL,
  "email" varchar NOT NULL,
  "secret_code" varchar NOT NULL,
  "is_used" bool NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "expired_at" timestamptz NOT NULL DEFAULT (now() + interval '15 minutes')
);

CREATE TABLE "accounts" (
  "id" bigserial PRIMARY KEY,
  "owner" varchar NOT NULL,
  "balance" bigint NOT NULL,
  "currency" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "entries" (
  "id" bigserial PRIMARY KEY,
  "account_id" bigint NOT NULL,
  "amount" bigint NOT NULL
);

CREATE TABLE "transfers" (
  "id" bigserial PRIMARY KEY,
  "from_account_id" bigint NOT NULL,
  "to_account_id" bigint NOT NULL,
  "amount" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "sessions" (
  "id" uuid PRIMARY KEY,
  "username" varchar NOT NULL,
  "refresh_token" varchar NOT NULL,
  "user_agent" varchar NOT NULL,
  "client_ip" varchar NOT NULL,
  "is_blocked" boolean NOT NULL DEFAULT false,
  "expires_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE INDEX "idx_accounts_owner" ON "accounts" ("owner");

CREATE UNIQUE INDEX "owner_currency_key" ON "accounts" ("owner", "currency");

CREATE INDEX "idx_entries_account_id" ON "entries" ("account_id");

CREATE INDEX "idx_transfers_from_account" ON "transfers" ("from_account_id");

CREATE INDEX "idx_transfers_to_account" ON "transfers" ("to_account_id");

CREATE INDEX "idx_transfers_from_to_account" ON "transfers" ("from_account_id", "to_account_id");

COMMENT ON TABLE "users" IS 'User accounts with authentication information';

COMMENT ON COLUMN "users"."username" IS 'Primary key - unique username';

COMMENT ON COLUMN "users"."hashed_password" IS 'Bcrypt hashed password';

COMMENT ON COLUMN "users"."full_name" IS 'User full name';

COMMENT ON COLUMN "users"."email" IS 'User email address';

COMMENT ON COLUMN "users"."is_email_verified" IS 'has the user email been verified?';

COMMENT ON COLUMN "users"."created_at" IS 'Account creation timestamp';

COMMENT ON COLUMN "users"."password_changed_at" IS 'Last password change timestamp';

COMMENT ON TABLE "accounts" IS 'Bank accounts belonging to users';

COMMENT ON COLUMN "accounts"."id" IS 'Auto-incrementing account ID';

COMMENT ON COLUMN "accounts"."owner" IS 'Account owner - references username';

COMMENT ON COLUMN "accounts"."balance" IS 'Account balance in smallest currency unit';

COMMENT ON COLUMN "accounts"."currency" IS 'Currency code (USD, EUR, etc.)';

COMMENT ON COLUMN "accounts"."created_at" IS 'Account creation timestamp';

COMMENT ON TABLE "entries" IS 'Account transaction entries (debits and credits)';

COMMENT ON COLUMN "entries"."id" IS 'Auto-incrementing entry ID';

COMMENT ON COLUMN "entries"."account_id" IS 'References account';

COMMENT ON COLUMN "entries"."amount" IS 'Transaction amount - can be negative or positive';

COMMENT ON TABLE "transfers" IS 'Money transfers between accounts';

COMMENT ON COLUMN "transfers"."id" IS 'Auto-incrementing transfer ID';

COMMENT ON COLUMN "transfers"."from_account_id" IS 'Source account';

COMMENT ON COLUMN "transfers"."to_account_id" IS 'Destination account';

COMMENT ON COLUMN "transfers"."amount" IS 'Transfer amount - must be positive';

COMMENT ON COLUMN "transfers"."created_at" IS 'Transfer timestamp';

COMMENT ON TABLE "sessions" IS 'User authentication sessions with refresh tokens';

COMMENT ON COLUMN "sessions"."id" IS 'Session UUID - matches refresh token ID';

COMMENT ON COLUMN "sessions"."username" IS 'Session owner';

COMMENT ON COLUMN "sessions"."refresh_token" IS 'Refresh token for authentication';

COMMENT ON COLUMN "sessions"."user_agent" IS 'Client user agent string';

COMMENT ON COLUMN "sessions"."client_ip" IS 'Client IP address';

COMMENT ON COLUMN "sessions"."is_blocked" IS 'Session blocked status';

COMMENT ON COLUMN "sessions"."expires_at" IS 'Session expiration time';

COMMENT ON COLUMN "sessions"."created_at" IS 'Session creation time';

ALTER TABLE "verify_emails" ADD FOREIGN KEY ("username") REFERENCES "users" ("username");

ALTER TABLE "accounts" ADD FOREIGN KEY ("owner") REFERENCES "users" ("username");

ALTER TABLE "entries" ADD FOREIGN KEY ("account_id") REFERENCES "accounts" ("id");

ALTER TABLE "transfers" ADD FOREIGN KEY ("from_account_id") REFERENCES "accounts" ("id");

ALTER TABLE "transfers" ADD FOREIGN KEY ("to_account_id") REFERENCES "accounts" ("id");

ALTER TABLE "sessions" ADD FOREIGN KEY ("username") REFERENCES "users" ("username");
