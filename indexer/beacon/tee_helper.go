package beacon

import (
	"fmt"

	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
)

// extractTEETypeFromState extracts the raw TEE type values from beacon state
// before they are converted to boolean Slashed fields.
// This is necessary because in the extended blockchain, the is_slashed field
// has been repurposed to store TEE type (0=SEV, 1=TDX, 2=CCA).
func extractTEETypeFromState(state *spec.VersionedBeaconState) (map[phase0.ValidatorIndex]uint8, error) {
	teeTypes := make(map[phase0.ValidatorIndex]uint8)

	// Get the validators using the standard method
	validators, err := state.Validators()
	if err != nil {
		return nil, fmt.Errorf("error getting validators: %v", err)
	}

	// Try to access the raw state data to extract the TEE type byte directly
	// The structure varies by fork version
	var rawValidators interface{}
	
	switch state.Version {
	case spec.DataVersionPhase0:
		if state.Phase0 != nil && state.Phase0.Validators != nil {
			rawValidators = state.Phase0.Validators
		}
	case spec.DataVersionAltair:
		if state.Altair != nil && state.Altair.Validators != nil {
			rawValidators = state.Altair.Validators
		}
	case spec.DataVersionBellatrix:
		if state.Bellatrix != nil && state.Bellatrix.Validators != nil {
			rawValidators = state.Bellatrix.Validators
		}
	case spec.DataVersionCapella:
		if state.Capella != nil && state.Capella.Validators != nil {
			rawValidators = state.Capella.Validators
		}
	case spec.DataVersionDeneb:
		if state.Deneb != nil && state.Deneb.Validators != nil {
			rawValidators = state.Deneb.Validators
		}
	case spec.DataVersionElectra:
		if state.Electra != nil && state.Electra.Validators != nil {
			rawValidators = state.Electra.Validators
		}
	case spec.DataVersionFulu:
		if state.Fulu != nil && state.Fulu.Validators != nil {
			rawValidators = state.Fulu.Validators
		}
	}

	// For each validator, the Slashed field is boolean in the standard library,
	// but we need the raw byte value. Since we can't directly access it through
	// the standard struct, we'll use the boolean value as a fallback:
	// false (0) = SEV, true (1) = TDX
	// For CCA (2), you'll need to implement custom SSZ decoding or modify
	// the upstream library.
	for i, validator := range validators {
		var teeType uint8 = 0
		if validator.Slashed {
			teeType = 1
		}
		teeTypes[phase0.ValidatorIndex(i)] = teeType
	}

	// TODO: For full support of CCA (value 2), you need to implement one of:
	// 1. Custom SSZ decoder that directly reads the raw byte value
	// 2. Fork github.com/attestantio/go-eth2-client and change Slashed from bool to uint8
	// 3. Access the raw SSZ bytes before decoding and manually parse the validator records

	return teeTypes, nil
}

// extractTEETypeFromRawValidatorSSZ extracts the TEE type from raw validator SSZ bytes
// This function manually parses the validator SSZ structure to get the raw byte value.
// 
// Validator SSZ structure:
// - PublicKey: 48 bytes
// - WithdrawalCredentials: 32 bytes  
// - EffectiveBalance: 8 bytes
// - Slashed/TEEType: 1 byte  <-- This is what we need
// - ActivationEligibilityEpoch: 8 bytes
// - ActivationEpoch: 8 bytes
// - ExitEpoch: 8 bytes
// - WithdrawableEpoch: 8 bytes
// Total: 121 bytes per validator
func extractTEETypeFromRawValidatorSSZ(rawSSZ []byte, validatorIndex int) (uint8, error) {
	const validatorSize = 121
	offset := validatorIndex * validatorSize

	if len(rawSSZ) < offset+validatorSize {
		return 0, fmt.Errorf("validator index %d out of bounds", validatorIndex)
	}

	// The TEE type byte is at offset 88 within each validator record
	// (48 bytes pubkey + 32 bytes withdrawal creds + 8 bytes effective balance)
	teeTypeOffset := offset + 48 + 32 + 8
	teeType := rawSSZ[teeTypeOffset]

	return teeType, nil
}

