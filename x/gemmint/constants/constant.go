package constants

import (
	"cosmossdk.io/math"
)

// Constants for validator emission
const (
	APR = "0.04"
)

var (
	BlocksPerYear = 365 * 24 * 60 * 60 / 5          // adjust if your block time is different
	ValidatorAPR  = math.LegacyNewDecWithPrec(4, 2) // 4% = 0.04
)

const (
	Denom            = "outbe"
	FeeCollectorName = "fee_collector"
	ModuleName       = "distribution"
	LegacyPrecision  = 18
)
