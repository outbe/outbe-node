package keeper

import (
	"context"
	"fmt"
	"strconv"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/outbe/outbe-node/x/distribution/constants"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	errortypes "github.com/outbe/outbe-node/errors"

	sdkerrors "cosmossdk.io/errors"
)

// AllocateTokens performs reward and fee distribution to all validators based
// on the F1 fee distribution specification.
func (k WrappedBaseKeeper) AllocateTokens(ctx context.Context, totalPreviousPower int64, bondedVotes []abci.VoteInfo) error {

	totalSelfBonded := sdkmath.ZeroInt()
	validators, _ := k.sk.GetAllValidators(ctx)

	for _, validator := range validators {
		if validator.IsJailed() || !validator.IsBonded() {
			return nil
		}
		totalSelfBonded = totalSelfBonded.Add(validator.GetTokens())
	}

	feeCollector := k.ak.GetModuleAccount(ctx, constants.FeeCollectorName)
	feesCollectedInt := k.bk.GetAllBalances(ctx, feeCollector.GetAddress())
	feesCollected := sdk.NewDecCoinsFromCoins(feesCollectedInt...)

	if !feesCollectedInt.IsValid() {
		return sdkerrors.Wrapf(errortypes.ErrInvalidCoins, "invalid fees collected: %s", feesCollectedInt.String())
	}

	if feesCollectedInt.IsZero() {
		return nil
	}

	if constants.FeeCollectorName == "" || constants.ModuleName == "" {
		return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "module names cannot be empty")
	}

	err := k.bk.SendCoinsFromModuleToModule(ctx, constants.FeeCollectorName, constants.ModuleName, feesCollectedInt)
	if err != nil {
		return sdkerrors.Wrapf(
			errortypes.ErrInsufficientFunds,
			"failed to transfer %s from %s to %s: %v",
			feesCollectedInt.String(),
			constants.FeeCollectorName,
			constants.ModuleName,
			err,
		)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := k.rk.GetParams(sdkCtx)

	rewardCoins := sdk.NewDecCoins()

	if feesCollected.IsValid() && !feesCollected.IsZero() {
		for _, validator := range validators {

			feeShare, err := k.rk.CalculateValidatorFeeShare(feesCollected[0].Amount, sdkmath.LegacyNewDecFromInt(validator.GetTokens()), sdkmath.LegacyNewDecFromInt(totalSelfBonded))
			if err != nil {
				return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "couldn't calculate validator fee share")
			}

			fmt.Println("0000000000000000000000000000-feesCollected[0].Amount", feesCollected[0].Amount)
			fmt.Println("0000000000000000000000000000-sdkmath.LegacyNewDecFromInt(validator.GetTokens()", sdkmath.LegacyNewDecFromInt(validator.GetTokens()))
			fmt.Println("0000000000000000000000000000-sdkmath.LegacyNewDecFromInt(totalSelfBonded)", sdkmath.LegacyNewDecFromInt(totalSelfBonded))
			fmt.Println("0000000000000000000000000000-sdkmath.LegacyMustNewDecFromStr(params.Apr)", sdkmath.LegacyMustNewDecFromStr(params.Apr))
			fmt.Println("0000000000000000000000000000-params.BlockPerYear", params.BlockPerYear)
			fmt.Println("0000000000000000000000000000-feeShare", feeShare)

			i, err := strconv.ParseInt(params.BlockPerYear, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to convert string '%s' to int64: %v", params.BlockPerYear, err)
			}

			minAprreward, err := k.rk.CalculateMinimumAPRReward(sdkmath.LegacyNewDecFromInt(validator.GetTokens()), sdkmath.LegacyMustNewDecFromStr(params.Apr), i)
			if err != nil {
				return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "failed to calculate min apr reward")
			}

			fmt.Println("11111111111111111111111111-minAprreward", minAprreward)

			if feeShare.GTE(minAprreward) {
				fmt.Println("here1 ---------------->")
				rewardCoins = rewardCoins.Add(sdk.NewDecCoin(constants.Denom, feeShare.TruncateInt()))

				valBz, err := k.sk.ValidatorAddressCodec().StringToBytes(validator.GetOperator())
				if err != nil {
					return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "validator address cannot be empty")
				}

				currentRewards, err := k.GetValidatorCurrentRewards(ctx, valBz)
				if err != nil {
					return sdkerrors.Wrapf(errortypes.ErrNotFound, "failed to get current rewards for validator %s: %v", string(valBz), err)
				}

				currentRewards.Rewards = currentRewards.Rewards.Add(rewardCoins...)
				err = k.SetValidatorCurrentRewards(ctx, valBz, currentRewards)
				if err != nil {
					return sdkerrors.Wrapf(errortypes.ErrNotFound, "failed to set current rewards for validator %s: %v", string(valBz), err)
				}
			} else {

				emissionNeeded := minAprreward.Sub(feeShare)
				fmt.Println("444444444444444444444444-emissionNeeded", emissionNeeded)
				if emissionNeeded.IsPositive() {
					fmt.Println("here2 ---------------->")
					emission := sdk.NewDecCoinFromDec(constants.Denom, sdkmath.LegacyNewDecWithPrec(emissionNeeded.TruncateInt64(), constants.LegacyPrecision))
					fmt.Println("emission ---------------->", emission)
					mintCoin := sdk.NewCoin(constants.Denom, emission.Amount.TruncateInt())
					mintCoins := sdk.NewCoins(mintCoin)
					fmt.Println("mintCoins ---------------->", mintCoins)
					if err := k.bk.MintCoins(ctx, types.ModuleName, mintCoins); err != nil {
						return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[AllocateTokens][MintCoins] failed. Failed to mint coins")
					}

					valBz, err := k.sk.ValidatorAddressCodec().StringToBytes(validator.GetOperator())
					if err != nil {
						return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "validator address cannot be empty")
					}

					currentRewards, err := k.GetValidatorCurrentRewards(ctx, valBz)
					if err != nil {
						return sdkerrors.Wrapf(errortypes.ErrNotFound, "failed to get current rewards for validator %s: %v", string(valBz), err)
					}
					fmt.Println("currentRewards ---------------->", currentRewards)
					currentRewards.Rewards = currentRewards.Rewards.Add(emission)
					err = k.SetValidatorCurrentRewards(ctx, valBz, currentRewards)
					if err != nil {
						return sdkerrors.Wrapf(errortypes.ErrNotFound, "failed to set current rewards for validator %s: %v", string(valBz), err)
					}
					fmt.Println("currentRewards --after set---------------->", currentRewards)
					sdkCtx.EventManager().EmitEvent(
						sdk.NewEvent(
							types.EventTypeRewards,
							sdk.NewAttribute(sdk.AttributeKeyAmount, mintCoin.String()),
							sdk.NewAttribute(types.AttributeKeyValidator, validator.GetOperator()),
						),
					)

					outstanding, err := k.GetValidatorOutstandingRewards(ctx, valBz)
					if err != nil {
						return err
					}
					fmt.Println("outstanding ---------------->", outstanding)
					outstanding.Rewards = outstanding.Rewards.Add(emission)
					k.SetValidatorOutstandingRewards(ctx, valBz, outstanding)

				}
			}
		}
	}
	return nil
}
