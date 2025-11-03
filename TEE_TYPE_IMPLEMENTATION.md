# TEE Type Implementation Summary

This document describes the changes made to display TEE (Trusted Execution Environment) type information on the validator detail page in Dora.

## Quick Summary

The validator detail page now displays the TEE type of each validator. The implementation stores TEE types in the database and displays them on the UI. **However, to fully support all three TEE types (0, 1, and 2) from incoming SSZ data, you'll need to implement Option A (forking go-eth2-client library)** as described below.

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

### Current Implementation

The implementation now:
- ✅ Stores TEE type values in cache and database
- ✅ Displays TEE types 0 (SEV), 1 (TDX), and 2 (CCA) correctly from the database
- ⚠️ Only captures values 0 and 1 from incoming SSZ data (due to `phase0.Validator.Slashed` being bool)

### SSZ Decoding Limitation

The `phase0.Validator` struct from the external library (`github.com/attestantio/go-eth2-client/spec/phase0`) defines `Slashed` as a `bool`, which can only represent two values (0 or 1). 

**For CCA (value 2) to work correctly**, you need to implement one of these solutions:

#### **Option A: Fork go-eth2-client Library (RECOMMENDED)**

This is the cleanest solution but requires maintaining a fork:

1. Fork `github.com/attestantio/go-eth2-client`
2. In `spec/phase0/validator.go`, change the `Slashed` field from `bool` to `uint8`:
   ```go
   type Validator struct {
       PublicKey                  BLSPubKey `ssz-size:"48"`
       WithdrawalCredentials      Root      `ssz-size:"32"`
       EffectiveBalance           Gwei      `ssz-size:"8"`
       Slashed                    uint8     `ssz-size:"1"`  // Changed from bool to uint8
       ActivationEligibilityEpoch Epoch     `ssz-size:"8"`
       ActivationEpoch            Epoch     `ssz-size:"8"`
       ExitEpoch                  Epoch     `ssz-size:"8"`
       WithdrawableEpoch          Epoch     `ssz-size:"8"`
   }
   ```
3. Update your `go.mod` to use the forked version:
   ```bash
   go mod edit -replace github.com/attestantio/go-eth2-client=github.com/yourusername/go-eth2-client@tee-support
   ```
4. Update the code that uses `Slashed` as bool to use it as uint8
5. Remove the conversion logic in `validatorcache.go`

**Pros:** Clean, type-safe, preserves original values  
**Cons:** Requires maintaining a fork, may need updates when upstream releases new versions

#### **Option B: Custom Raw SSZ Parser**

Implement a custom HTTP client that fetches raw SSZ bytes and parses the validators list manually:

1. Add a new method `GetStateRawSSZ()` to `clients/consensus/rpc/beaconapi.go`
2. Fetch the raw SSZ response from the beacon API
3. Manually parse the validators list using `protolambda/ztyp` or similar
4. Extract the TEE type byte at offset 88 within each validator
5. Store these values before calling the standard `GetState()` method

**Pros:** No fork required, full control over parsing  
**Cons:** Complex implementation, need to maintain SSZ structure knowledge

#### **Option C: Modify Your Beacon Node**

If you control the beacon node implementation:

1. Add a custom API endpoint that returns raw TEE type values
2. Call this endpoint separately to get TEE types for validators
3. Merge the data when populating the validator cache

**Note:** If validators with CCA (value 2) already exist in your database, they will display correctly - the limitation only affects **new** validators being decoded from incoming SSZ data.

## Testing

After applying the changes and running the migration:

1. Start the Dora explorer
2. Navigate to a validator detail page: `http://localhost:51092/validator/<validator_pubkey_or_index>`
3. Verify that the TEE Type field appears between Status and Balance
4. Check that it displays the correct TEE type for your validators

## Files Modified

1. `db/schema/pgsql/20251101000001_tee-type-field.sql` (new migration)
2. `db/schema/sqlite/20251101000001_tee-type-field.sql` (new migration)
3. `dbtypes/dbtypes.go` - Changed `Validator.Slashed` from `bool` to `uint8`
4. `types/models/validator.go` - Added `ValidatorPageData.TeeType` field
5. `handlers/validator.go` - Retrieves TEE type from DB and displays on page
6. `indexer/beacon/validatorcache.go` - Added `teeType` cache field, extracts and stores TEE type
7. `templates/validator/validator.html` - Displays TEE type row

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

