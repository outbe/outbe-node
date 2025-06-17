package bindings

// QCorechainQuery contains QCorechain custom queries.
type QcoreQuery struct {
	QueryBlockEmissionRequest *QueryBlockEmissionRequest `json:"query_block_emission_request,omitempty"`
}
type QueryBlockEmissionRequest struct {
	BlockNumber string `json:"block_number"`
}
type QueryBlockEmissionResponse struct {
	BlockEmission string `json:"block_emission"`
}
