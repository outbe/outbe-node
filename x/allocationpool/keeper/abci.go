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

		emission, found := k.GetTotalEmission(ctx)
		if !found {
			emissionToken, err := k.CalculateExponentialBlockEmission(ctx, 1)
			if err != nil {
				logger.Error("Failed to calculate exponential block emission for block: 1",
					"error", err)
				return sdkerrors.Wrap(err, "[CalculateExponentialBlockEmission] failed to calculate exponential block emission")
			}
			emission := types.Emission{
				BlockNumber:       strBlockNumber,
				TotalEmission:     emissionToken,
				EmissionTimestamp: strTimeStamp,
			}
			k.SetEmission(ctx, emission)

			logger.Info("✅ Completed exponential token emission for block 1 — emission stored in context",
				"total_emission", emissionToken,
				"block_number", ctx.BlockHeight())

			return nil
		}

		decEmission, err := sdkmath.LegacyNewDecFromStr(emission.TotalEmission)
		if err != nil {
			logger.Error("Failed to convert total emission to sdk.Dec",
				"error", err,
				"TotalEmission", emission.TotalEmission)
			return sdkerrors.Wrap(err, "failed to convert total emission to sdk.Dec")
		}

		inflation := k.mintKeeper.GetAllMinters(ctx)[0].Inflation

		if inflation.GT(sdkmath.LegacyNewDecWithPrec(2, 2)) {

			emissionToken, err := k.CalculateExponentialBlockEmission(ctx, ctx.BlockHeight())
			if err != nil {
				logger.Error("failed to calculate exponential block emission",
					"error", err,
					"emission_token", emissionToken)
				return sdkerrors.Wrap(err, "failed to calculate exponential block emission")
			}

			decEmissionPerBlock, err := sdkmath.LegacyNewDecFromStr(emissionToken)
			if err != nil {
				logger.Error("Failed to convert emission token to sdk.Dec",
					"error", err,
					"emission_token", emissionToken)
				return sdkerrors.Wrap(err, "failed to convert emission token to sdk.Dec")
			}

			emission.BlockNumber = strBlockNumber
			emission.TotalEmission = decEmission.Add(decEmissionPerBlock).String()
			emission.EmissionTimestamp = strTimeStamp

			k.SetEmission(ctx, emission)

			return nil

		} else {

			emissionToken, err := k.CalculateFixedBlockEmission(ctx)
			if err != nil {
				logger.Error("[CalculateFixedBlockEmission] failed to calculate fixed block emission",
					"error", err,
					"emission_token", emissionToken)
				return sdkerrors.Wrap(err, "failed to calculate fixed block emission")
			}

			decEmissionPerBlock, err := sdkmath.LegacyNewDecFromStr(emissionToken)
			if err != nil {
				ctx.Logger().Error("Failed to convert emission token to sdk.Dec",
					"error", err,
					"emission_token", emissionToken)
				return sdkerrors.Wrap(err, "failed to convert emission token to sdk.Dec")
			}

			emission.BlockNumber = strBlockNumber
			emission.TotalEmission = decEmission.Add(decEmissionPerBlock).String()
			emission.EmissionTimestamp = strTimeStamp

			k.SetEmission(ctx, emission)

			return nil
		}
	}

	return nil
}
