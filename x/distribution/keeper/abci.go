package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// BeginBlocker will persist the current header and validator set as a historical entry
// and prune the oldest entry based on the HistoricalEntries parameter
func (k WrappedBaseKeeper) BeginBlocker(ctx context.Context) error {
	defer telemetry.ModuleMeasureSince(stakingtypes.ModuleName, telemetry.Now(), telemetry.MetricKeyBeginBlocker)

	var previousTotalPower int64

	wctx := sdk.UnwrapSDKContext(ctx)

	if err := k.AllocateTokens(ctx, previousTotalPower, wctx.VoteInfos()); err != nil {
		return err
	}
	return nil
}
