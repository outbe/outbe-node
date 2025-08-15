package keeper

import (
	"strconv"
	"time"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/outbe/outbe-node/x/allocationpool/types"
)

func (k Keeper) BeginBlocker(ctx sdk.Context) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	logger := k.Logger(ctx)

	if ctx.BlockHeight() > 0 {

		strBlockNumber := strconv.FormatInt(ctx.BlockHeight(), 10)
		strTimeStamp := strconv.FormatInt(ctx.BlockTime().Unix(), 10)

		inflation := k.mintKeeper.GetAllMinters(ctx)[0].Inflation
		if inflation.GT(sdkmath.LegacyNewDecWithPrec(2, 2)) {
			emissionToken, err := k.CalculateExponentialBlockEmission(ctx, ctx.BlockHeight())
			if err != nil {
				logger.Error("failed to calculate exponential block emission",
					"error", err,
					"emission_token", emissionToken)
				return sdkerrors.Wrap(err, "failed to calculate exponential block emission")
			}

			newEmission := types.Emission{
				BlockNumber:         strBlockNumber,
				ActualEmission:      emissionToken,
				RemainBlockEmission: emissionToken,
				EmissionTimestamp:   strTimeStamp,
			}
			k.SetEmission(ctx, newEmission)

			return nil

		} else {
			emissionToken, err := k.CalculateFixedBlockEmission(ctx)
			if err != nil {
				logger.Error("[CalculateFixedBlockEmission] failed to calculate fixed block emission",
					"error", err,
					"emission_token", emissionToken)
				return sdkerrors.Wrap(err, "failed to calculate fixed block emission")
			}

			newEmission := types.Emission{
				BlockNumber:         strBlockNumber,
				ActualEmission:      emissionToken,
				RemainBlockEmission: emissionToken,
				EmissionTimestamp:   strTimeStamp,
			}
			k.SetEmission(ctx, newEmission)

			return nil
		}
	}

	return nil
}
