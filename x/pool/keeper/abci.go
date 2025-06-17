package keeper

import (
	"strconv"
	"time"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/outbe/outbe-node/errors"
	"github.com/outbe/outbe-node/x/pool/constants"
	"github.com/outbe/outbe-node/x/pool/types"
)

func (k Keeper) BeginBlocker(ctx sdk.Context) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	if ctx.BlockHeight() > 0 {

		strBlockNumber := strconv.FormatInt(ctx.BlockHeight(), 10)
		strTimeStamp := strconv.FormatInt(ctx.BlockTime().Unix(), 10)

		emission, found := k.GetTotalEmission(ctx)
		if !found {
			emissionToken, err := k.CalculateExponentialBlockEmission(1)
			if err != nil {
				return sdkerrors.Wrap(errortypes.ErrCalculation, "[BeginBlocker] failed. CalculateExponentialTokens failed")
			}
			emission := types.Emission{
				BlockNumber:       strBlockNumber,
				TotalEmission:     emissionToken,
				EmissionTimestamp: strTimeStamp,
			}
			k.SetEmission(ctx, emission)
			return nil
		}

		decEmission, err := sdkmath.LegacyNewDecFromStr(emission.TotalEmission)
		if err != nil {
			return sdkerrors.Wrap(errortypes.ErrInvalidType, "[BeginBlocker] failed. couldn't convert string to sdk.Dec")
		}

		if ctx.BlockHeight() < constants.TransitionBlockNumber {
			emissionToken, err := k.CalculateExponentialBlockEmission(ctx.BlockHeight())
			if err != nil {
				return sdkerrors.Wrap(errortypes.ErrCalculation, "[BeginBlocker] failed. CalculateExponentialTokens failed")
			}

			decEmissionPerBlock, err := sdkmath.LegacyNewDecFromStr(emissionToken)
			if err != nil {
				return sdkerrors.Wrap(errortypes.ErrInvalidType, "[BeginBlocker] failed. failed to convert string to sdk.Dec")
			}

			emission.BlockNumber = strBlockNumber
			emission.TotalEmission = decEmission.Add(decEmissionPerBlock).String()
			emission.EmissionTimestamp = strTimeStamp

			k.SetEmission(ctx, emission)
			return nil
		} else {
			emissionToken, err := k.CalculateFixedBlockEmission(ctx)
			if err != nil {
				return sdkerrors.Wrap(errortypes.ErrCalculation, "[BeginBlocker] failed. CalculateFixedTokens failed")
			}

			decEmissionPerBlock, err := sdkmath.LegacyNewDecFromStr(emissionToken)
			if err != nil {
				return sdkerrors.Wrap(errortypes.ErrInvalidType, "[BeginBlocker[ failed. failed to convert string to sdk.Dec")
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
