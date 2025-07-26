package keeper

import (
	"context"
	"math"
	"math/big"

	sdkmath "cosmossdk.io/math"
	"github.com/outbe/outbe-node/app/params"
	errortypes "github.com/outbe/outbe-node/errors"
	"github.com/outbe/outbe-node/x/allocationpool/constants"
	"github.com/outbe/outbe-node/x/allocationpool/types"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SetEmission(ctx context.Context, emission types.Emission) error {
	store := k.storeService.OpenKVStore(ctx)
	b := k.cdc.MustMarshal(&emission)
	return store.Set(types.GetEmissionKey("pool_emission"), b)
}

func (k Keeper) GetTotalEmission(ctx context.Context) (val types.Emission, found bool) {
	store := k.storeService.OpenKVStore(ctx)
	emissionKey := types.GetEmissionKey("pool_emission")
	b, err := store.Get(emissionKey)

	if b == nil || err != nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

func (k Keeper) CalculateExponentialBlockEmission(ctx context.Context, blockNumber int64) (string, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	logger := k.Logger(sdkCtx)

	logger.Info("🔁 Starting exponential token emission for allocation pool — inflation increases up to 2% per block")

	initialRate, err := sdkmath.LegacyNewDecFromStr(k.GetParams(ctx).InitialRate)
	if err != nil {
		logger.Error("Failed to parse initial rate", "error", err, "block_height", sdkCtx.BlockHeight())
		return "", sdkerrors.Wrap(errortypes.ErrInvalidCoins, "[CalculateExponentialBlockEmission] failed to parse initial rate")
	}

	decay, err := sdkmath.LegacyNewDecFromStr(k.GetParams(ctx).Decay)
	if err != nil {
		logger.Error("Failed to parse decay rate",
			"error", err,
			"block_height", sdkCtx.BlockHeight())
		return "", sdkerrors.Wrap(err, "[CalculateExponentialBlockEmission] failed to parse decay rate")
	}

	n := sdkmath.LegacyNewDec(blockNumber)
	decayN := decay.Mul(n)

	expArg := -decayN.MustFloat64()
	expResult := math.Exp(expArg)

	scaled := new(big.Float).SetFloat64(expResult)
	scaled.Mul(scaled, big.NewFloat(1e18))
	scaledInt := new(big.Int)
	scaled.Int(scaledInt)

	expVal := sdkmath.LegacyNewDecFromBigInt(scaledInt)
	expVal = expVal.QuoInt64(1e18)

	tokens := initialRate.Mul(expVal)

	logger.Info("✅ Exponential block emission successfully calculated",
		"initial_rate", initialRate,
		"decay", decay,
		"block_number", blockNumber,
		"decay_n", decayN,
		"exp_val", expVal,
		"tokens", tokens,
	)

	return tokens.String(), nil
}

func (k Keeper) CalculateFixedBlockEmission(ctx context.Context) (string, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	logger := k.Logger(sdkCtx)

	logger.Info("🔁 Starting exponential token emission for allocation pool — inflation decreases down to 2% per block")

	// Fixed emission: (totalSupply * 0.02) / 365 / 17280
	emissionRate, err := sdkmath.LegacyNewDecFromStr(constants.EmissionRate)
	if err != nil {
		return "", sdkerrors.Wrapf(errortypes.ErrInvalidCoins, "[CalculateFixedBlockEmission:] failed to parse EmissionRate: %v", err)
	}

	daysPerYear, err := sdkmath.LegacyNewDecFromStr(constants.DaysPerYear)
	if err != nil {
		return "", sdkerrors.Wrapf(errortypes.ErrInvalidCoins, "[CalculateFixedBlockEmission:] failed to parse DaysPerYear: %v", err)
	}

	blocksPerDay, err := sdkmath.LegacyNewDecFromStr(constants.BlocksPerDay)
	if err != nil {
		return "", sdkerrors.Wrapf(errortypes.ErrInvalidCoins, "[CalculateFixedBlockEmission:] failed to parse BlocksPerDay: %v", err)
	}

	totalSupply := k.bankKeeper.GetSupply(ctx, params.BondDenom)
	if totalSupply.IsNil() || totalSupply.IsZero() || totalSupply.IsNegative() {
		return "", sdkerrors.Wrap(errortypes.ErrInvalidCoins, "[CalculateFixedBlockEmission:] total supply must be greater than zero and valid")
	}

	decTotalSupply := sdkmath.LegacyNewDecFromInt(totalSupply.Amount)
	emissionPerBlock := decTotalSupply.Mul(emissionRate).Quo(daysPerYear).Quo(blocksPerDay)

	logger.Info("✅ Fixed block emission successfully calculated",
		"total_supply", totalSupply,
		"emission_rate", emissionRate,
		"days_per_year", daysPerYear,
		"blocks_per_day", blocksPerDay,
		"dec_total_supply", decTotalSupply,
		"emission_per_block", emissionPerBlock,
	)

	return emissionPerBlock.String(), nil
}
