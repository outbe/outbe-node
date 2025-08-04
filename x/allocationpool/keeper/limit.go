package keeper

import (
	"context"

	sdkmath "cosmossdk.io/math"
	"github.com/outbe/outbe-node/app/params"
	"github.com/outbe/outbe-node/x/allocationpool/constants"
)

func (k Keeper) CalculateAnnualEmissionLimit(ctx context.Context) string {
	totalSupply := k.bankKeeper.GetSupply(ctx, params.BondDenom)

	limitRate := sdkmath.LegacyMustNewDecFromStr(constants.LimitRate) // 2%
	supply := sdkmath.LegacyNewDecFromInt(totalSupply.Amount)         // total supply

	// annual_limit = supply * 0.02
	annualLimit := supply.Mul(limitRate)

	return annualLimit.String()
}
