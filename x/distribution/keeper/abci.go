package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	errortypes "github.com/outbe/outbe-node/errors"
)

// BeginBlocker will persist the current header and validator set as a historical entry
// and prune the oldest entry based on the HistoricalEntries parameter
func (k WrappedBaseKeeper) BeginBlocker(ctx context.Context) error {
	defer telemetry.ModuleMeasureSince(stakingtypes.ModuleName, telemetry.Now(), telemetry.MetricKeyBeginBlocker)

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	logger := k.Logger(sdkCtx)

	wctx := sdk.UnwrapSDKContext(ctx)

	// determine the total power signing the block
	var previousTotalPower int64

	validators, err := k.sk.GetAllValidators(ctx)
	if err != nil {
		return sdkerrors.Wrapf(errortypes.ErrInvalidType,
			"[BeginBlocker][Distribution] failed to query all validators: %v", err)
	}

	if err := k.AllocateTokens(ctx, previousTotalPower, validators); err != nil {
		logger.Error("[BeginBlocker][Distribution] failed to distribute validator rewards", "error", err.Error())
		return sdkerrors.Wrapf(err, "[BeginBlocker][Distribution] failed to distribute transactions fee for a validator at block %d", sdkCtx.BlockHeight())
	}

	for _, voteInfo := range wctx.VoteInfos() {
		previousTotalPower += voteInfo.Validator.Power
	}

	if wctx.BlockHeight() > 1 {
		if err := k.AllocateBlockProvisioningTokens(ctx, previousTotalPower, wctx.VoteInfos()); err != nil {
			return sdkerrors.Wrapf(err, "[BeginBlocker][Distribution] failed to distribute block provisioning rewards for a validator at block %d", sdkCtx.BlockHeight())
		}
	}

	// record the proposer for when we payout on the next block
	consAddr := sdk.ConsAddress(wctx.BlockHeader().ProposerAddress)
	return k.SetPreviousProposerConsAddr(ctx, consAddr)
}
