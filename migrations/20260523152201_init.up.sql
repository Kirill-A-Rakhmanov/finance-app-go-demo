CREATE TYPE "transaction_type" AS ENUM ('income', 'expense');
CREATE TYPE "transaction_category_type" AS ENUM ('income', 'expense');

CREATE TABLE "users" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "first_name" character varying(100) NOT NULL,
  "last_name" character varying(100) NOT NULL,
  "middle_name" character varying(100) NULL,
  "email" character varying(255) NOT NULL,
  "password" character varying(255) NOT NULL,
  PRIMARY KEY ("id")
);
CREATE TABLE "accounts" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "user_id" uuid NOT NULL,
  "name" character varying(255) NOT NULL,
  "icon" character varying(255) NOT NULL,
  "currency" character varying(3) NOT NULL,
  "is_negative_balance_enabled" boolean NOT NULL DEFAULT false,
  "is_hidden" boolean NOT NULL DEFAULT false,
  "is_excluded_from_count" boolean NOT NULL DEFAULT false,
  "adjustment" numeric(19,4) NOT NULL DEFAULT 0.00,
  PRIMARY KEY ("id")
);
CREATE TABLE "transaction_categories" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "user_id" uuid NOT NULL,
  "parent_id" uuid NULL,
  "name" character varying(100) NOT NULL,
  "slug" character varying(100) NULL,
  "icon" character varying(100) NOT NULL,
  "color" character varying(100) NOT NULL,
  "is_default" boolean NOT NULL DEFAULT false,
  "type" "transaction_category_type" NOT NULL,
  PRIMARY KEY ("id")
);
CREATE TABLE "counterparties" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "user_id" uuid NOT NULL,
  "name" character varying(255) NOT NULL,
  PRIMARY KEY ("id")
);
CREATE TABLE "transfers" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "from_account_id" uuid NOT NULL,
  "to_account_id" uuid NOT NULL,
  "from_amount" numeric(19,4) NOT NULL,
  "to_amount" numeric(19,4) NOT NULL,
  "date" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
CREATE TABLE "transactions" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "account_id" uuid NOT NULL,
  "transaction_category_id" uuid NULL,
  "counterparty_id" uuid NULL,
  "transfer_id" uuid NULL,
  "type" "transaction_type" NOT NULL,
  "amount" numeric(19,4) NOT NULL,
  "name" character varying(255) NULL,
  "date" timestamptz NOT NULL,
  "is_confirmed" boolean NOT NULL DEFAULT false,
  PRIMARY KEY ("id")
);
CREATE TABLE "feature_flags" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "key" character varying(100) NOT NULL,
  "description" character varying(255) NOT NULL DEFAULT '',
  "enabled" boolean NOT NULL DEFAULT false,
  PRIMARY KEY ("id")
);
CREATE TABLE "feature_flag_overrides" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "feature_flag_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "enabled" boolean NOT NULL,
  PRIMARY KEY ("id")
);

ALTER TABLE "users"
    ADD CONSTRAINT "uni_users_email"
    UNIQUE ("email");

ALTER TABLE "accounts"
ADD CONSTRAINT "fk_users_accounts"
    FOREIGN KEY ("user_id")
    REFERENCES "users" ("id")
    ON UPDATE NO ACTION ON DELETE CASCADE;

CREATE INDEX "idx_accounts_user_id"
    ON "accounts" ("user_id");

ALTER TABLE "transaction_categories"
    ADD CONSTRAINT "fk_users_transaction_categories"
    FOREIGN KEY ("user_id")
    REFERENCES "users" ("id")
    ON UPDATE NO ACTION ON DELETE CASCADE;

ALTER TABLE "transaction_categories"
    ADD CONSTRAINT "fk_transaction_categories_parent"
    FOREIGN KEY ("parent_id")
    REFERENCES "transaction_categories" ("id")
    ON UPDATE NO ACTION ON DELETE CASCADE;

CREATE UNIQUE INDEX "uni_transaction_categories_user_id_type_name"
    ON "transaction_categories" ("user_id", "type", "name");

CREATE INDEX "idx_transaction_categories_parent_id"
    ON "transaction_categories" ("parent_id")
    WHERE "parent_id" IS NOT NULL;

ALTER TABLE "counterparties"
    ADD CONSTRAINT "fk_users_counterparties"
    FOREIGN KEY ("user_id")
    REFERENCES "users" ("id")
    ON UPDATE NO ACTION ON DELETE CASCADE;

CREATE UNIQUE INDEX "uni_counterparties_user_id_name"
    ON "counterparties" ("user_id", "name");

ALTER TABLE "transfers"
    ADD CONSTRAINT "fk_accounts_from_transfers"
    FOREIGN KEY ("from_account_id")
    REFERENCES "accounts" ("id")
    ON UPDATE NO ACTION ON DELETE CASCADE;

ALTER TABLE "transfers"
    ADD CONSTRAINT "fk_accounts_to_transfers"
    FOREIGN KEY ("to_account_id")
    REFERENCES "accounts" ("id")
    ON UPDATE NO ACTION ON DELETE CASCADE;

CREATE INDEX "idx_transfers_from_account_id"
    ON "transfers" ("from_account_id");

CREATE INDEX "idx_transfers_to_account_id"
    ON "transfers" ("to_account_id");

ALTER TABLE "transactions"
    ADD CONSTRAINT "fk_accounts_transactions"
    FOREIGN KEY ("account_id")
    REFERENCES "accounts" ("id")
    ON UPDATE NO ACTION ON DELETE CASCADE;

ALTER TABLE "transactions"
    ADD CONSTRAINT "fk_transaction_categories_transactions"
    FOREIGN KEY ("transaction_category_id")
    REFERENCES "transaction_categories" ("id")
    ON UPDATE NO ACTION ON DELETE SET NULL;

ALTER TABLE "transactions"
    ADD CONSTRAINT "fk_counterparties_transactions"
    FOREIGN KEY ("counterparty_id")
    REFERENCES "counterparties" ("id")
    ON UPDATE NO ACTION ON DELETE SET NULL;

ALTER TABLE "transactions"
    ADD CONSTRAINT "fk_transactions_transfer"
    FOREIGN KEY ("transfer_id")
    REFERENCES "transfers" ("id")
    ON UPDATE NO ACTION ON DELETE CASCADE;

CREATE INDEX "idx_transactions_account_id_date"
    ON "transactions" ("account_id", "date" DESC);

CREATE INDEX "idx_transactions_transaction_category_id"
    ON "transactions" ("transaction_category_id")
    WHERE "transaction_category_id" IS NOT NULL;

CREATE INDEX "idx_transactions_counterparty_id"
    ON "transactions" ("counterparty_id")
    WHERE "counterparty_id" IS NOT NULL;

CREATE INDEX "idx_transactions_transfer_id"
    ON "transactions" ("transfer_id")
    WHERE "transfer_id" IS NOT NULL;

ALTER TABLE "feature_flags"
    ADD CONSTRAINT "uni_feature_flags_key"
    UNIQUE ("key");

ALTER TABLE "feature_flag_overrides"
    ADD CONSTRAINT "fk_feature_flags_overrides"
    FOREIGN KEY ("feature_flag_id")
    REFERENCES "feature_flags" ("id")
    ON UPDATE NO ACTION ON DELETE CASCADE;

ALTER TABLE "feature_flag_overrides"
    ADD CONSTRAINT "fk_users_feature_flag_overrides"
    FOREIGN KEY ("user_id")
    REFERENCES "users" ("id")
    ON UPDATE NO ACTION ON DELETE CASCADE;

CREATE UNIQUE INDEX "uni_feature_flag_overrides_feature_flag_id_user_id"
    ON "feature_flag_overrides" ("feature_flag_id", "user_id");

CREATE INDEX "idx_feature_flag_overrides_user_id"
    ON "feature_flag_overrides" ("user_id");
