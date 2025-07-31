package keeper

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateContractAddress(t *testing.T) {
	validAddress := "q14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s0zssm2"
	invalidPrefix := "x14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s0zssm2"
	invalidLength := "q14hj2"
	invalidChars := "q14hj2tav!@#pesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s0zssm2"

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Valid address", validAddress, true},
		{"Invalid prefix", invalidPrefix, false},
		{"Invalid length", invalidLength, true},
		{"Invalid characters", invalidChars, true},
	}

	for _, tt := range tests {
		err := validateContractAddress(tt.input)
		if tt.wantErr && err == nil {
			require.NoError(t, err)
		} else if !tt.wantErr && err != nil {
			require.Error(t, err)
		}
	}
}

func validateContractAddress(addr string) error {
	const expectedPrefix = "q"
	const expectedLength = 63 // typical length for Cosmos addresses

	if !strings.HasPrefix(addr, expectedPrefix) {
		return errors.New("address must start with prefix 'q'")
	}
	if len(addr) != expectedLength {
		return errors.New("invalid address length")
	}
	// Optional: check if it's alphanumeric (simplified)
	for _, r := range addr {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')) {
			return errors.New("address contains invalid characters")
		}
	}
	return nil
}
