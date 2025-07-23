package types

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/outbe/outbe-node/app/params"
)

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
		Period: &Period{
			CurrentPeriod:     0,
			PeriodStartHeight: 0,
			CurrentSeed:       nil,
			InCommitPhase:     true,
			CommitEndHeight:   0,
			RevealEndHeight:   0,
		},
		Commitments: []*Commitment{},
		Reveals:     []*Reveal{},
	}
}

// TODPO these values comes fron the genesis
// DefaultParams returns the default parameters
func DefaultParams() Params {
	return Params{
		CommitPeriod:   10,
		RevealPeriod:   10,
		MinimumDeposit: &sdk.Coin{Denom: params.BondDenom, Amount: math.NewInt(1000)},
		PenaltyAmount:  &sdk.Coin{Denom: params.BondDenom, Amount: math.NewInt(500)},
	}
}

// Validate validates the genesis state
func (gs GenesisState) Validate() error {
	if gs.Params.CommitPeriod == 0 {
		return fmt.Errorf("commit period cannot be zero")
	}
	if gs.Params.RevealPeriod == 0 {
		return fmt.Errorf("reveal period cannot be zero")
	}
	if !gs.Params.MinimumDeposit.IsValid() {
		return fmt.Errorf("invalid minimum deposit: %s", gs.Params.MinimumDeposit)
	}
	if !gs.Params.PenaltyAmount.IsValid() {
		return fmt.Errorf("invalid penalty amount: %s", gs.Params.PenaltyAmount)
	}
	return nil
}
