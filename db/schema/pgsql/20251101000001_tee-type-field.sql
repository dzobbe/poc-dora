-- +goose Up
-- +goose StatementBegin

-- Alter slashed column to SMALLINT to support TEE type (0=SEV, 1=TDX, 2=CCA)
ALTER TABLE public."validators" 
    ALTER COLUMN slashed TYPE SMALLINT USING (CASE WHEN slashed THEN 1 ELSE 0 END);

COMMENT ON COLUMN public."validators".slashed IS 'TEE Type: 0=SEV, 1=TDX, 2=CCA';

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin

-- Revert to BOOLEAN
ALTER TABLE public."validators" 
    ALTER COLUMN slashed TYPE BOOLEAN USING (slashed > 0);

COMMENT ON COLUMN public."validators".slashed IS NULL;

-- +goose StatementEnd

