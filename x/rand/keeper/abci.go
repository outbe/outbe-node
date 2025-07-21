package keeper

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/outbe/outbe-node/x/rand/types"
)

func (k Keeper) BeginBlocker(ctx sdk.Context) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	state, _ := k.GetPeriod(ctx)
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

	return nil
}

func (k Keeper) EndBlocker(ctx sdk.Context) error {
	state, _ := k.GetPeriod(ctx)
	currentHeight := ctx.BlockHeight()

	if !state.InCommitPhase && currentHeight >= state.RevealEndHeight {
		newRandomness := k.GenerateRandomness(ctx, state.CurrentPeriod)

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

	validators, _ := k.stakingKeeper.GetAllValidators(ctx)
	fmt.Println("validators------------------->", validators)

	// k.UpdateRandParticipants(ctx, validators, state.CurrentPeriod)

	return nil
}
