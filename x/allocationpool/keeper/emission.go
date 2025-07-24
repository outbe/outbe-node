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
	initialRate, err := sdkmath.LegacyNewDecFromStr(k.GetParams(ctx).InitialRate)
	if err != nil {
		return "", sdkerrors.Wrapf(errortypes.ErrInvalidCoins, "[CalculateExponentialBlockEmission][LegacyNewDecFromStr] failed. err:%s", err)
	}

	decay, err := sdkmath.LegacyNewDecFromStr(k.GetParams(ctx).Decay)
	if err != nil {
		return "", sdkerrors.Wrapf(errortypes.ErrInvalidCoins, "[CalculateExponentialBlockEmission][LegacyNewDecFromStr] failed. err:%s", err)
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

	return tokens.String(), nil
}

func (k Keeper) CalculateFixedBlockEmission(ctx context.Context) (string, error) {

	// Fixed emission: (totalSupply * 0.02) / 365 / 17280
	apr, err := sdkmath.LegacyNewDecFromStr(constants.APR)
	if err != nil {
		return "", sdkerrors.Wrapf(errortypes.ErrInvalidCoins, "[CalculateFixedDailyEmission][LegacyNewDecFromStr] failed. err:%s", err)
	}

	daysPerYear, err := sdkmath.LegacyNewDecFromStr(constants.DaysPerYear)
	if err != nil {
		return "", sdkerrors.Wrapf(errortypes.ErrInvalidCoins, "[CalculateFixedDailyEmission][LegacyNewDecFromStr] failed. err:%s", err)
	}

	blocksPerDay, err := sdkmath.LegacyNewDecFromStr(constants.BlocksPerDay)
	if err != nil {
		return "", sdkerrors.Wrapf(errortypes.ErrInvalidCoins, "[CalculateFixedDailyEmission][LegacyNewDecFromStr] failed. err:%s", err)
	}

	totalSupply := k.bankKeeper.GetSupply(ctx, params.BondDenom)
	if totalSupply.IsNil() || totalSupply.IsZero() || totalSupply.IsNegative() {
		return "", sdkerrors.Wrap(errortypes.ErrInvalidCoins, "[CalculateFixedDailyEmission][TotalSupplyAll] failed. total supply must be greater than zero")
	}

	decTotalSupply := sdkmath.LegacyNewDecFromInt(totalSupply.Amount)
	emissionPerBlock := decTotalSupply.Mul(apr).Quo(daysPerYear).Quo(blocksPerDay)

	return emissionPerBlock.String(), nil
}

func (k Keeper) CalculateFixedDailyEmission(ctx context.Context) (string, error) {

	// Fixed emission: (totalSupply * 0.02) / 365
	apr, err := sdkmath.LegacyNewDecFromStr(constants.APR)
	if err != nil {
		return "", sdkerrors.Wrapf(errortypes.ErrInvalidCoins, "[CalculateFixedDailyEmission][LegacyNewDecFromStr] failed. err:%s", err)
	}

	daysPerYear, err := sdkmath.LegacyNewDecFromStr(constants.DaysPerYear)
	if err != nil {
		return "", err
	}

	totalSupply := k.TotalSupplyAll(ctx)
	if totalSupply[0].TotalSupply <= "0" {
		return "", sdkerrors.Wrap(errortypes.ErrInvalidCoins, "[CalculateFixedDailyEmission][TotalSupplyAll] failed. total supply must be greater than zero.")

	}

	decTotalSupply := sdkmath.LegacyMustNewDecFromStr(totalSupply[0].TotalSupply)
	dailyEmissionPerBlock := decTotalSupply.Mul(apr).Quo(daysPerYear)

	return dailyEmissionPerBlock.String(), nil
}

func (k Keeper) CalculateFixedAnnualEmission(ctx context.Context) (string, error) {

	// Fixed emission: (totalSupply * 0.02)
	apr, err := sdkmath.LegacyNewDecFromStr(constants.APR)
	if err != nil {
		return "", sdkerrors.Wrapf(errortypes.ErrInvalidCoins, "[CalculateFixedAnnualEmission][LegacyNewDecFromStr] failed. err:%s", err)
	}

	totalSupply := k.TotalSupplyAll(ctx)

	decTotalSupply := sdkmath.LegacyMustNewDecFromStr(totalSupply[0].TotalSupply)
	annualEmissionPerBlock := decTotalSupply.Mul(apr)

	return annualEmissionPerBlock.String(), nil
}
