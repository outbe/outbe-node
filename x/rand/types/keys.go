package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	ModuleName   = "rand"
	StoreKey     = ModuleName
	RouterKey    = ModuleName
	QuerierRoute = ModuleName
)

// Store key prefixes
var (
	PeriodKey     = []byte{0x01}
	CommitmentKey = []byte{0x02}
	RevealKey     = []byte{0x03}
	ParamsKey     = []byte{0x04}
	PenaltyKey    = []byte{0x05}
)

func GetPeriodKey(id string) []byte {
	return append(PeriodKey, address.MustLengthPrefix([]byte(id))...)
}

// GetCommitmentKey constructs the key for a commitment
func GetCommitmentKey(period uint64, validator string) []byte {
	return append(CommitmentKey, append(sdk.Uint64ToBigEndian(period), []byte(validator)...)...)
}

// GetRevealKey constructs the key for a reveal
func GetRevealKey(period uint64, validator string) []byte {
	return append(RevealKey, append(sdk.Uint64ToBigEndian(period), []byte(validator)...)...)
}

func GetPenaltyKey(period uint64, validator sdk.ValAddress) []byte {
	return append(PenaltyKey, append(sdk.Uint64ToBigEndian(period), validator.Bytes()...)...)
}
