package constants

// Constants for validator emission
const (
	APR = "0.04"
)

var (
	BlocksPerYear = 365 * 24 * 60 * 60 / 5 // adjust if your block time is different
)

const (
	Denom            = "unit"
	FeeCollectorName = "fee_collector"
	ModuleName       = "distribution"
	LegacyPrecision  = 18
)

const (
	ExpectedPrefix = "outbe"
	ExpectedLength = 64 // typical length for Cosmos addresses
)
