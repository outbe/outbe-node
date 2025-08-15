package types

import (
	fmt "fmt"

	sdkmath "cosmossdk.io/math"
)

// NewRewardParams creates a new RewardParams instance.
func NewRewardParams() Params {
	return Params{
		ThetaCra:         sdkmath.LegacyNewDecWithPrec(4, 2),
		PerCraCap:        sdkmath.LegacyNewDecWithPrec(32, 2),
		EmaLambda:        sdkmath.LegacyNewDecWithPrec(4, 1),
		BetaMin:          sdkmath.LegacyNewDecWithPrec(6, 1),
		BetaMax:          sdkmath.LegacyNewDecWithPrec(15, 1),
		AnomalyThreshold: sdkmath.LegacyNewDecWithPrec(4, 1),
		BBase:            sdkmath.LegacyDec(sdkmath.NewInt(1000)),
		BScale:           sdkmath.LegacyNewDec(1),
		GracePeriodDays:  100,
	}
}

// DefaultParams returns the default parameters for the allocationpool module.
func DefaultParams() Params {
	return NewRewardParams()
}

// Validate validates the parameters.
func (p Params) Validate() error {
	if p.ThetaCra.IsNegative() || p.ThetaCra.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("invalid ThetaCRA")
	}
	// TODO : Similarly validate others...
	return nil
}
