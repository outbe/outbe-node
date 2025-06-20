package reward

import (
	"context"

	"github.com/cosmos/cosmos-sdk/telemetry"
	"github.com/outbe/outbe-node/x/reward/keeper"
	"github.com/outbe/outbe-node/x/reward/types"
)

func BeginBlocker(ctx context.Context, k keeper.Keeper) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, telemetry.Now(), telemetry.MetricKeyBeginBlocker)

	// err := k.Distribution(ctx)
	// if err != nil {
	// 	fmt.Println("555555555555555error", err)
	// }
	return nil
}
