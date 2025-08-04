package bindings

// OutbeQuery contains Outbe chain custom queries.
type OutbeQuery struct {
	QueryBlockEmissionRequest *QueryBlockEmissionRequest `json:"query_block_emission_request,omitempty"`
}
type QueryBlockEmissionRequest struct {
	BlockNumber string `json:"block_number"`
}
type QueryBlockEmissionResponse struct {
	BlockEmission string `json:"block_emission"`
}
