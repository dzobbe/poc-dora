# TEE Type Implementation Summary

This document describes the changes made to display TEE (Trusted Execution Environment) type information on the validator detail page in Dora.

## Overview

In the extended blockchain, the `is_slashed` field in the validator SSZ data has been repurposed to store the TEE type, which can have the following values:
- `0` = SEV (AMD Secure Encrypted Virtualization)
- `1` = TDX (Intel Trust Domain Extensions)  
- `2` = CCA (Arm Confidential Compute Architecture)

## Changes Made

### 1. Database Schema Updates

**New Migration Files:**
- `db/schema/pgsql/20251101000001_tee-type-field.sql`
- `db/schema/sqlite/20251101000001_tee-type-field.sql`

These migrations alter the `slashed` column in the `validators` table from `BOOLEAN` to `SMALLINT/INTEGER` to support values 0, 1, and 2.

**To apply the migration:**
```bash
# Run your migration tool (e.g., goose)
goose -dir db/schema/pgsql up
# or
goose -dir db/schema/sqlite up
```

### 2. Type Definition Updates

**File: `dbtypes/dbtypes.go`**
- Changed `Validator.Slashed` field type from `bool` to `uint8`
- Added comment explaining it's repurposed as TEE type

**File: `types/models/validator.go`**
- Added `TeeType` field (type `uint8`) to `ValidatorPageData` struct

### 3. Handler Updates

**File: `handlers/validator.go`**
- Added import for `db` package
- Retrieves TEE type value directly from database using `db.GetValidatorByIndex()`
- Populates the `TeeType` field in the page data

### 4. Data Conversion Logic

**File: `indexer/beacon/validatorcache.go`**

Two conversion functions were updated:

1. **`UnwrapDbValidator()`**: Converts database validator (with uint8 TEE type) to `phase0.Validator` (with bool Slashed)
   - `0` → `false`
   - `>0` → `true`

2. **Validator persistence**: Converts `phase0.Validator` (bool) to database format (uint8)
   - `false` → `0`
   - `true` → `1`

### 5. Template Updates

**File: `templates/validator/validator.html`**
- Added new row to display TEE Type after Status field
- Shows decoded value: "SEV", "TDX", "CCA", or "Unknown (value)" for unexpected values

## Important Notes

### SSZ Decoding Limitation

The `phase0.Validator` struct from the external library (`github.com/attestantio/go-eth2-client/spec/phase0`) defines `Slashed` as a `bool`, which can only represent two values (0 or 1). 

**For CCA (value 2) to work correctly**, you may need to:

1. **Option A**: Modify your SSZ decoding to directly access the raw byte value and store it correctly in the database with value 2.

2. **Option B**: Fork the `go-eth2-client` library and change the `Slashed` field type to `uint8` in the `phase0.Validator` struct.

3. **Option C**: Create a custom validator struct that extends `phase0.Validator` with an additional field for the raw TEE type value.

Currently, the implementation will:
- Correctly display all three TEE types if the value is stored in the database as 0, 1, or 2
- But SSZ data with value 2 will be decoded as `true` (equivalent to value 1) when using the standard library

## Testing

After applying the changes and running the migration:

1. Start the Dora explorer
2. Navigate to a validator detail page: `http://localhost:51092/validator/<validator_pubkey_or_index>`
3. Verify that the TEE Type field appears between Status and Balance
4. Check that it displays the correct TEE type for your validators

## Files Modified

1. `db/schema/pgsql/20251101000001_tee-type-field.sql` (new)
2. `db/schema/sqlite/20251101000001_tee-type-field.sql` (new)
3. `dbtypes/dbtypes.go`
4. `types/models/validator.go`
5. `handlers/validator.go`
6. `indexer/beacon/validatorcache.go`
7. `templates/validator/validator.html`

## Next Steps

1. **Apply database migration** to update the `slashed` column type
2. **Test the implementation** with your extended blockchain
3. **If using CCA (value 2)**, verify that the SSZ decoding properly stores value 2 in the database
4. **Optional**: Add TEE type filtering to the validators list page
5. **Optional**: Add TEE type information to the API endpoints

## Rollback

If you need to rollback these changes:

```bash
# Rollback the migration
goose -dir db/schema/pgsql down
# or  
goose -dir db/schema/sqlite down
```

Then revert the code changes in the modified files.

