-- +goose Up
-- +goose StatementBegin

-- SQLite doesn't support ALTER COLUMN TYPE directly, so we need to recreate the table
CREATE TABLE IF NOT EXISTS "validators_new" (
    validator_index BIGINT NOT NULL,
    pubkey BLOB NOT NULL,
    withdrawal_credentials BLOB NOT NULL,
    effective_balance BIGINT NOT NULL,
    slashed INTEGER NOT NULL, -- TEE Type: 0=SEV, 1=TDX, 2=CCA
    activation_eligibility_epoch BIGINT NOT NULL,
    activation_epoch BIGINT NOT NULL,
    exit_epoch BIGINT NOT NULL,
    withdrawable_epoch BIGINT NOT NULL,
    PRIMARY KEY (validator_index)
);

-- Copy data, converting BOOLEAN to INTEGER
INSERT INTO "validators_new" 
SELECT 
    validator_index,
    pubkey,
    withdrawal_credentials,
    effective_balance,
    CAST(slashed AS INTEGER) as slashed,
    activation_eligibility_epoch,
    activation_epoch,
    exit_epoch,
    withdrawable_epoch
FROM "validators";

-- Drop old table and rename new one
DROP TABLE "validators";
ALTER TABLE "validators_new" RENAME TO "validators";

-- Recreate index
CREATE INDEX IF NOT EXISTS "validators_pubkey_idx"
    ON "validators" ("pubkey");

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin

-- Recreate table with BOOLEAN
CREATE TABLE IF NOT EXISTS "validators_new" (
    validator_index BIGINT NOT NULL,
    pubkey BLOB NOT NULL,
    withdrawal_credentials BLOB NOT NULL,
    effective_balance BIGINT NOT NULL,
    slashed BOOLEAN NOT NULL,
    activation_eligibility_epoch BIGINT NOT NULL,
    activation_epoch BIGINT NOT NULL,
    exit_epoch BIGINT NOT NULL,
    withdrawable_epoch BIGINT NOT NULL,
    PRIMARY KEY (validator_index)
);

-- Copy data back, converting INTEGER to BOOLEAN
INSERT INTO "validators_new" 
SELECT 
    validator_index,
    pubkey,
    withdrawal_credentials,
    effective_balance,
    CASE WHEN slashed > 0 THEN 1 ELSE 0 END as slashed,
    activation_eligibility_epoch,
    activation_epoch,
    exit_epoch,
    withdrawable_epoch
FROM "validators";

-- Drop old table and rename new one
DROP TABLE "validators";
ALTER TABLE "validators_new" RENAME TO "validators";

-- Recreate index
CREATE INDEX IF NOT EXISTS "validators_pubkey_idx"
    ON "validators" ("pubkey");

-- +goose StatementEnd

