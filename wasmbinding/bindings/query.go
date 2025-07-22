package bindings

// GemchainQuery contains Gemchain custom queries.
type OutbeQuery struct {
	Randomness *Randomness `json:"randomness,omitempty"`
}

type Randomness struct {
}

type RandomnessResponse struct {
	Period     string `json:"period"`
	Randomness string `json:"randomness"`
}
