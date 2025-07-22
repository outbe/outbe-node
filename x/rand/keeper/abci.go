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

func (k Keeper) BeginBlocker(ctx sdk.Context) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	logger := k.Logger(ctx)

	state, found := k.GetPeriod(ctx)
	if !found {
		logger.Error("[GetPeriod] Failed to get period state",
			"found", found,
		)
		return sdkerrors.Wrap(errortypes.ErrInvalidPhase, "failed to get period state")
	}
	currentHeight := ctx.BlockHeight()

	if state.InCommitPhase && currentHeight >= state.CommitEndHeight {
		state.InCommitPhase = false
		k.SetPeriod(ctx, state)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeRevealPhaseStart,
				sdk.NewAttribute(types.AttributeKeyPeriodNumber, fmt.Sprintf("%d", state.CurrentPeriod)),
				sdk.NewAttribute(types.AttributeKeyRevealEndHeight, fmt.Sprintf("%d", state.RevealEndHeight)),
			),
		)
	}

	logger.Info("Transitioned to reveal phase",
		"period", state.CurrentPeriod,
		"reveal_end_height", state.RevealEndHeight,
	)

	return nil
}

func (k Keeper) EndBlocker(ctx sdk.Context) error {
	logger := k.Logger(ctx)

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

		k.SetPeriod(ctx, state)
		k.PenalizeNonRevealers(ctx, state.CurrentPeriod-1)
		k.ClearPeriodData(ctx, state.CurrentPeriod-1)

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
	}

	logger.Info("New period started",
		"period", state.CurrentPeriod,
	)

	return nil
}
