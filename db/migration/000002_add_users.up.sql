-- add users table
CREATE TABLE
    "users" (
        "username" varchar PRIMARY KEY,
        "hashed_password" varchar NOT NULL,
        "full_name" varchar NOT NULL,
        "email" varchar NOT NULL UNIQUE,
        "created_at" timestamptz NOT NULL DEFAULT now (),
        "password_changed_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
    );

-- one account per currency for a given user
ALTER TABLE "accounts" ADD CONSTRAINT "owner_currency_key" UNIQUE ("owner", "currency");

ALTER TABLE "accounts" ADD FOREIGN KEY ("owner") REFERENCES "users" ("username");