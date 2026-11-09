DROP INDEX IF EXISTS "idx_feature_flag_overrides_user_id";
DROP INDEX IF EXISTS "idx_transactions_transfer_id";
DROP INDEX IF EXISTS "idx_transactions_counterparty_id";
DROP INDEX IF EXISTS "idx_transactions_transaction_category_id";
DROP INDEX IF EXISTS "idx_transactions_account_id_date";
DROP INDEX IF EXISTS "idx_transfers_to_account_id";
DROP INDEX IF EXISTS "idx_transfers_from_account_id";
DROP INDEX IF EXISTS "idx_transaction_categories_parent_id";
DROP INDEX IF EXISTS "idx_accounts_user_id";

DROP INDEX IF EXISTS "uni_feature_flag_overrides_feature_flag_id_user_id";

ALTER TABLE IF EXISTS "feature_flag_overrides"
    DROP CONSTRAINT IF EXISTS "fk_users_feature_flag_overrides";

ALTER TABLE IF EXISTS "feature_flag_overrides"
    DROP CONSTRAINT IF EXISTS "fk_feature_flags_overrides";

ALTER TABLE IF EXISTS "feature_flags"
    DROP CONSTRAINT IF EXISTS "uni_feature_flags_key";

ALTER TABLE IF EXISTS "transactions"
    DROP CONSTRAINT IF EXISTS "fk_transactions_transfer";

ALTER TABLE IF EXISTS "transactions"
    DROP CONSTRAINT IF EXISTS "fk_counterparties_transactions";

ALTER TABLE IF EXISTS "transactions"
    DROP CONSTRAINT IF EXISTS "fk_transaction_categories_transactions";

ALTER TABLE IF EXISTS "transactions"
    DROP CONSTRAINT IF EXISTS "fk_accounts_transactions";

ALTER TABLE IF EXISTS "transfers"
    DROP CONSTRAINT IF EXISTS "fk_accounts_to_transfers";

ALTER TABLE IF EXISTS "transfers"
    DROP CONSTRAINT IF EXISTS "fk_accounts_from_transfers";

ALTER TABLE IF EXISTS "counterparties"
    DROP CONSTRAINT IF EXISTS "fk_users_counterparties";

DROP INDEX IF EXISTS "uni_counterparties_user_id_name";
DROP INDEX IF EXISTS "uni_transaction_categories_user_id_type_name";

ALTER TABLE IF EXISTS "transaction_categories"
    DROP CONSTRAINT IF EXISTS "fk_users_transaction_categories";

ALTER TABLE IF EXISTS "accounts"
DROP CONSTRAINT IF EXISTS "fk_users_accounts";

ALTER TABLE IF EXISTS "users"
    DROP CONSTRAINT IF EXISTS "uni_users_email";

DROP TABLE IF EXISTS "feature_flag_overrides";
DROP TABLE IF EXISTS "feature_flags";
DROP TABLE IF EXISTS "transactions";
DROP TABLE IF EXISTS "counterparties";
DROP TABLE IF EXISTS "transfers";
DROP TABLE IF EXISTS "transaction_categories";
DROP TABLE IF EXISTS "accounts";
DROP TABLE IF EXISTS "users";

DROP TYPE IF EXISTS "transaction_type";
DROP TYPE IF EXISTS "transaction_category_type";
