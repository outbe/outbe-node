package keeper

import (
	"encoding/hex"
	"fmt"
	"time"

	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/outbe/outbe-node/errors"
	"github.com/outbe/outbe-node/x/rand/types"
)

// BeginBlocker handles the logic at the beginning of each block
func (k Keeper) BeginBlocker(ctx sdk.Context) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	logger := k.Logger(ctx)

	// Get period state with error handling
	state, found := k.GetPeriod(ctx)
	if !found {
		logger.Error("[GetPeriod] Failed to get period state",
			"found", found,
		)
		return sdkerrors.Wrap(errortypes.ErrInvalidPhase, "failed to get period state")
	}

	currentHeight := ctx.BlockHeight()

	logger.Info("BeginBlocker started", "height", currentHeight, "period", state.CurrentPeriod, "in_commit_phase", state.InCommitPhase)

	if state.InCommitPhase && currentHeight >= state.CommitEndHeight {
		state.InCommitPhase = false
		if err := k.SetPeriod(ctx, state); err != nil {
			logger.Error("Failed to update period state",
				"error", err,
			)
			return sdkerrors.Wrap(errortypes.ErrInvalidPhase, "failed to update period state")
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeRevealPhaseStart,
				sdk.NewAttribute(types.AttributeKeyPeriodNumber, fmt.Sprintf("%d", state.CurrentPeriod)),
				sdk.NewAttribute(types.AttributeKeyRevealEndHeight, fmt.Sprintf("%d", state.RevealEndHeight)),
			),
		)
		logger.Info("Transitioned to reveal phase",
			"period", state.CurrentPeriod,
			"reveal_end_height", state.RevealEndHeight,
		)
	}
	return nil
}

// EndBlocker handles the logic at the end of each block
func (k Keeper) EndBlocker(ctx sdk.Context) error {
	logger := k.Logger(ctx)

	// Get period state with error handling
	state, found := k.GetPeriod(ctx)
	if !found {
		logger.Error("Failed to get period state",
			"error", found,
		)
		return sdkerrors.Wrap(errortypes.ErrInvalidPhase, "failed to get period state")
	}

	currentHeight := ctx.BlockHeight()

	logger.Info("EndBlocker started",
		"height", currentHeight,
		"period", state.CurrentPeriod,
		"in_commit_phase", state.InCommitPhase,
	)

	// Handle reveal phase completion
	if !state.InCommitPhase && currentHeight >= state.RevealEndHeight {
		newRandomness, err := k.GenerateRandomness(ctx, state.CurrentPeriod)
		if newRandomness == nil || err != nil {
			logger.Error("Failed to generate randomness",
				"period", state.CurrentPeriod,
			)
			return sdkerrors.Wrapf(errortypes.ErrInvalidPhase, "failed to generate randomness for period %d", state.CurrentPeriod)
		}

		params := k.GetParams(ctx)
		state.CurrentPeriod++
		state.Block = uint64(currentHeight)
		state.CurrentSeed = newRandomness
		state.InCommitPhase = true
		state.PeriodStartHeight = currentHeight
		state.CommitEndHeight = currentHeight + int64(params.CommitPeriod)
		state.RevealEndHeight = state.CommitEndHeight + int64(params.RevealPeriod)

		if err := k.SetPeriod(ctx, state); err != nil {
			logger.Error("Failed to update period state",
				"error", err,
			)
			return sdkerrors.Wrapf(errortypes.ErrInvalidPhase, "failed to update period state: %s", err)
		}

		if err := k.PenalizeNonRevealers(ctx, state.CurrentPeriod-1); err != nil {
			logger.Error("Failed to penalize non-revealers", "period", state.CurrentPeriod-1, "error", err)
			sdkerrors.Wrapf(errortypes.ErrInvalidPhase, "failed to penalize non-revealers: %s", err)
		}

		k.ClearPeriodData(ctx, state.CurrentPeriod-1)
		if err := k.ClearPeriodData(ctx, state.CurrentPeriod-1); err != nil {
			logger.Error("Failed to clear period data",
				"period", state.CurrentPeriod-1,
				"error", err,
			)
			sdkerrors.Wrapf(errortypes.ErrInvalidPhase, "failed to clear period data: %s", err)
		}

		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.EventTypeRandomnessGenerated,
				sdk.NewAttribute(types.AttributeKeyPeriodNumber, fmt.Sprintf("%d", state.CurrentPeriod-1)),
				sdk.NewAttribute(types.AttributeKeyRandomness, hex.EncodeToString(newRandomness)),
			),
			sdk.NewEvent(
				types.EventTypeEpochStart,
				sdk.NewAttribute(types.AttributeKeyPeriodNumber, fmt.Sprintf("%d", state.CurrentPeriod)),
				sdk.NewAttribute(types.AttributeKeyCommitEndHeight, fmt.Sprintf("%d", state.CommitEndHeight)),
				sdk.NewAttribute(types.AttributeKeyRevealEndHeight, fmt.Sprintf("%d", state.RevealEndHeight)),
			),
		})

		logger.Info("New period started",
			"period", state.CurrentPeriod,
		)
	}

	// Handle commit phase
	if state.InCommitPhase {
		params := k.GetParams(ctx)
		validators, err := k.stakingKeeper.GetAllValidators(ctx)
		if err != nil {
			logger.Error("Failed to get validators",
				"error", err,
			)
			return sdkerrors.Wrapf(errortypes.ErrInvalidPhase, "failed to get validators: %s", err)
		}

		for _, validator := range validators {
			valAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
			if err != nil {
				logger.Error("failed to parse validator address",
					"validator", validator.OperatorAddress,
					"error", err)
				return sdkerrors.Wrap(errortypes.ErrInvalidAddress, "failed to parse validator address")
			}

			delegations, err := k.stakingKeeper.GetValidatorDelegations(ctx, sdk.ValAddress(valAddr))
			if err != nil {
				logger.Error("failed to get validator delegations",
					"validator", validator.OperatorAddress,
					"error", err)
				return sdkerrors.Wrap(errortypes.ErrInvalidState, "failed to get validator delegations")
			}

			logger.Info("successfully retrieved validator delegations",
				"validator", validator.OperatorAddress,
				"delegation_count", len(delegations),
				"delegations", delegations[0].DelegatorAddress,
			)

			if validator.IsBonded() {
				validatorAddr := validator.GetOperator()
				commitments := k.GetValidatorCommitmentByPeriod(ctx, state.CurrentPeriod, validatorAddr)

				logger.Debug("Processing validator",
					"address", validatorAddr,
					"commitments", len(commitments),
				)

				if len(commitments) == 0 {
					r := k.GenerateRandomValueBinary(ctx, sdk.ValAddress(validatorAddr), state.CurrentPeriod)
					if r == nil {
						logger.Error("Failed to generate random value",
							"validator", validatorAddr,
							"period", state.CurrentPeriod,
						)
						continue // Skip to next validator instead of failing entire block
					}

					k.WriteByteToTheFile(ctx, r)

					computedHash := k.ComputeHash(r)
					if computedHash == nil {
						logger.Error("failed to compute hash", "validator",
							validatorAddr,
							"period", state.CurrentPeriod,
						)
						continue
					}

					err := k.Commit(ctx, state.CurrentPeriod, validatorAddr, *params.MinimumDeposit, computedHash, delegations[0].DelegatorAddress)
					if err != nil {
						logger.Error("failed to commit",
							"validator", validatorAddr,
							"period", state.CurrentPeriod,
							"error", err,
						)
						continue // Continue processing other validators
					}

					logger.Info("Commitment successful",
						"validator", validatorAddr,
						"period", state.CurrentPeriod,
					)
				}
			}
		}
	}

	// Handle reveal phase
	if !state.InCommitPhase {
		validators, err := k.stakingKeeper.GetAllValidators(ctx)
		if err != nil {
			logger.Error("failed to get all validators",
				"period", state.CurrentPeriod,
				"error", err)
			return sdkerrors.Wrap(errortypes.ErrInvalidState, "failed to get validators")
		}

		for _, validator := range validators {
			if validator.IsBonded() {
				valAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
				if err != nil {
					logger.Error("failed to parse validator address",
						"validator", validator.OperatorAddress,
						"error", err)
					return sdkerrors.Wrap(errortypes.ErrInvalidAddress, "failed to parse validator address")
				}

				delegations, err := k.stakingKeeper.GetValidatorDelegations(ctx, sdk.ValAddress(valAddr))
				if err != nil {
					logger.Error("failed to get validator delegations",
						"validator", validator.OperatorAddress,
						"error", err)
					return sdkerrors.Wrap(errortypes.ErrInvalidState, "failed to get validator delegations")
				}

				validatorAddr := validator.GetOperator()
				err = k.Reveal(ctx, validatorAddr, state.CurrentPeriod, delegations[0].DelegatorAddress)
				if err != nil {
					logger.Error("failed to reveal",
						"validator", validatorAddr,
						"period", state.CurrentPeriod,
						"error", err)
					continue // Continue processing other validators
				}

				logger.Info("Reveal successful",
					"validator", validatorAddr,
					"period", state.CurrentPeriod,
				)
			}
		}
	}

	return nil
}
