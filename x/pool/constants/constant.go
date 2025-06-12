package constants

// Constants for exponential token emission
const (
	InitialRate = "18475.316174578344" // Initial rate in tokens per block
	Decay       = "0.00000005"         // Decay rate (5×10^-8)
)

// Constants for fixed token emission
const (
	TransitionBlockNumber = 17472000
	APR                   = "0.02" // Annual inflation rate
	DaysPerYear           = "365"
	BlocksPerDay          = "17280" // Blocks per day
	// TotalSupply           = "10000000"
)
