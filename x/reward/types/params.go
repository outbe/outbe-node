package types

// NewRewardParams creates a new RewardParams instance.
func NewRewardParams(initialRate, decay string) Params {
	return Params{
		Apr:              "",
		BlockPerYear:     "",
		MaxSelfBondToken: "",
	}
}

// DefaultParams returns the default parameters for the allocationpool module.
func DefaultParams() Params {
	return NewRewardParams("65536", "0.00000006")
}

// Validate validates the parameters.
func (p Params) Validate() error {
	// if p.InitialRate == "" {
	// 	return fmt.Errorf("initial_rate cannot be empty")
	// }
	// if p.Decay == "" {
	// 	return fmt.Errorf("decay cannot be empty")
	// }
	// // Add additional validation if needed (e.g., check if InitialRate is a valid integer)
	// _, found := math.NewIntFromString(p.InitialRate)
	// if !found {
	// 	return fmt.Errorf("invalid initial_rate: %w", found)
	// }
	// _, err := math.NewDecFromString(p.Decay)
	// if err != nil {
	// 	return fmt.Errorf("invalid decay: %w", found)
	// }
	return nil
}
