package keeper

import (
	"context"

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

	err := k.bk.SendCoinsFromModuleToModule(ctx, constants.FeeCollectorName, constants.ModuleName, feesCollectedInt)
	if err != nil {
		return err
	}

	rewardCoins := sdk.NewDecCoins()

	if feesCollected.IsValid() && !feesCollected.IsZero() {
		for _, validator := range validators {
			feeShare := feesCollected[0].Amount.Mul(sdkmath.LegacyNewDecFromInt(validator.GetTokens())).Quo(sdkmath.LegacyNewDecFromInt(totalSelfBonded))

			x := (sdkmath.LegacyMustNewDecFromStr(constants.Apr).Quo(sdkmath.LegacyMustNewDecFromStr(constants.BlockPerYear)))
			minAprReward := sdkmath.LegacyNewDecFromInt(validator.GetTokens()).Mul(x)

			if feeShare.GTE(minAprReward) {
				rewardCoins = rewardCoins.Add(sdk.NewDecCoin(constants.Denom, feeShare.TruncateInt()))
				//
				valBz, err := k.sk.ValidatorAddressCodec().StringToBytes(validator.GetOperator())
				if err != nil {
					return err
				}
				//
				currentRewards, err := k.GetValidatorCurrentRewards(ctx, valBz)
				if err != nil {
					return err
				}

				currentRewards.Rewards = currentRewards.Rewards.Add(rewardCoins[0])
				err = k.SetValidatorCurrentRewards(ctx, valBz, currentRewards)
				if err != nil {
					return err
				}
			} else {
				emissionNeeded := minAprReward.Sub(sdkmath.LegacyNewDecFromInt(feeShare.TruncateInt())).TruncateInt()
				if emissionNeeded.IsPositive() {
					emission := sdk.NewDecCoin(constants.Denom, emissionNeeded)

					mintCoin := sdk.NewCoin(constants.Denom, emissionNeeded)
					mintCoins := sdk.NewCoins(mintCoin)
					if err := k.bk.MintCoins(ctx, types.ModuleName, mintCoins); err != nil {
						return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[AllocateTokens][MintCoins] failed. Failed to mint coins")
					}

					valBz, err := k.sk.ValidatorAddressCodec().StringToBytes(validator.GetOperator())
					if err != nil {
						return err
					}

					currentRewards, err := k.GetValidatorCurrentRewards(ctx, valBz)
					if err != nil {
						return err
					}

					currentRewards.Rewards = currentRewards.Rewards.Add(emission)
					err = k.SetValidatorCurrentRewards(ctx, valBz, currentRewards)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
