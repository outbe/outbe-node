package types

const (
	EventTypeEpochStart          = "epoch_start"
	EventTypeCommitment          = "commitment"
	EventTypeReveal              = "reveal"
	EventTypeRevealPhaseStart    = "reveal_phase_start"
	EventTypeValidatorSetChanged = "validator_set_changed"
	EventTypePenalty             = "penalty"
	EventTypeRandomnessGenerated = "randomness_generated"

	AttributeKeyPeriodNumber    = "period_number"
	AttributeKeyCommitEndHeight = "commit_end_height"
	AttributeKeyRevealEndHeight = "reveal_end_height"
	AttributeKeyValidator       = "validator"
	AttributeKeyRandomness      = "randomness"
)
