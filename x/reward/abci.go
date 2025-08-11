package reward

import (
	"context"

	"github.com/cosmos/cosmos-sdk/telemetry"
	"github.com/outbe/outbe-node/x/reward/keeper"
	"github.com/outbe/outbe-node/x/reward/types"
)

func BeginBlocker(ctx context.Context, k keeper.Keeper) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, telemetry.Now(), telemetry.MetricKeyBeginBlocker)

	return nil
}
