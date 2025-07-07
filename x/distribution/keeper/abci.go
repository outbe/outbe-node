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

	wctx := sdk.UnwrapSDKContext(ctx)

	// determine the total power signing the block
	var previousTotalPower int64

	for _, voteInfo := range wctx.VoteInfos() {
		previousTotalPower += voteInfo.Validator.Power
	}

	if err := k.AllocateTokens(ctx, previousTotalPower, wctx.VoteInfos()); err != nil {
		return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "failed to distribute validator rewards")
	}

	// record the proposer for when we payout on the next block
	consAddr := sdk.ConsAddress(wctx.BlockHeader().ProposerAddress)
	return k.SetPreviousProposerConsAddr(ctx, consAddr)
}
