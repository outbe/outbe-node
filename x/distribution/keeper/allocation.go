package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/outbe/outbe-node/app/params"
	errortypes "github.com/outbe/outbe-node/errors"
	"github.com/outbe/outbe-node/x/distribution/constants"
)

func (k WrappedBaseKeeper) AllocateTokens(ctx context.Context, totalPreviousPower int64, bondedVotes []abci.VoteInfo) error {
	// Fetch and clear collected fees from the previous block
	feeCollector := k.ak.GetModuleAccount(ctx, constants.FeeCollectorName)
	feesCollectedInt := k.bk.GetAllBalances(ctx, feeCollector.GetAddress())
	feesCollected := sdk.NewDecCoinsFromCoins(feesCollectedInt...)

	if feesCollected.IsZero() || !feesCollected.IsValid() || feesCollected.Empty() {
		return nil
	}

	// Transfer collected fees to the distribution module account
	err := k.bk.SendCoinsFromModuleToModule(ctx, constants.FeeCollectorName, types.ModuleName, feesCollectedInt)
	if err != nil {
		return err
	}

	// Get fee pool for community pool updates
	feePool, err := k.FeePool.Get(ctx)
	if err != nil {
		return err
	}

	// Handle case with no validators
	if len(bondedVotes) == 0 {
		feePool.CommunityPool = feePool.CommunityPool.Add(feesCollected...)
		return k.FeePool.Set(ctx, feePool)
	}

	// Calculate total self-bonded tokens across all validators
	totalSelfBonded, err := k.rk.CalculateTotalSelfBondedTokens(ctx, bondedVotes)
	if err != nil {
		return err
	}

	if totalSelfBonded.IsZero() || totalSelfBonded.IsNil() {
		feePool.CommunityPool = feePool.CommunityPool.Add(feesCollected...)
		return k.FeePool.Set(ctx, feePool)
	}

	// Initialize remaining fees and total emission
	remaining := feesCollected
	var totalEmission sdk.DecCoins

	// Get community tax (assumed 10%)
	communityTax, err := k.GetCommunityTax(ctx)
	if err != nil {
		return err
	}
	if communityTax.IsNil() {
		communityTax = sdkmath.LegacyNewDecWithPrec(1, 1)
	}

	// Distribute rewards to each validator
	for _, vote := range bondedVotes {
		validator, err := k.sk.ValidatorByConsAddr(ctx, vote.Validator.Address)
		if err != nil {
			return err
		}

		// Get validator's self-bonded tokens
		selfBonded, err := k.rk.GetValidatorSelfBondedTokens(ctx, validator)
		if err != nil {
			return err
		}

		var validatorFeeShare sdk.DecCoins
		for _, fee := range feesCollected {
			if fee.Denom == params.BondDenom {
				share := k.rk.CalculateFeeShare(fee.Amount, selfBonded, totalSelfBonded)
				validatorFeeShare = validatorFeeShare.Add(sdk.NewDecCoinFromDec(fee.Denom, share))
			}
		}

		minAprRewardAmount := k.rk.CalculateMinApr(ctx, selfBonded)
		minAprReward := sdk.NewDecCoins(sdk.NewDecCoinFromDec(params.BondDenom, minAprRewardAmount))

		// Determine reward and emission
		var reward sdk.DecCoins
		var emission sdk.DecCoins

		if validatorFeeShare[0].Amount.GT(minAprReward[0].Amount) {
			reward = validatorFeeShare
		} else {
			// Fee share is insufficient, use minimum APR reward
			reward = minAprReward
			for _, minReward := range minAprReward {
				feeShareCoin := validatorFeeShare.AmountOf(minReward.Denom)
				if feeShareCoin.IsPositive() {
					emissionAmount := minReward.Amount.Sub(feeShareCoin)
					if emissionAmount.IsPositive() {
						emission = emission.Add(sdk.NewDecCoinFromDec(minReward.Denom, emissionAmount))
					}
				} else {
					emission = emission.Add(minReward)
				}
			}
		}

		// Mint emission if needed
		if !emission.IsZero() {
			// Mint DecCoins directly to preserve fractional amounts
			err = k.bk.MintCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(params.BondDenom, emission[0].Amount.RoundInt())))
			if err != nil {
				return err
			}
			totalEmission = totalEmission.Add(emission...)
		}

		// Allocate reward to validator
		err = k.AllocateTokensToValidator(ctx, validator, reward)
		if err != nil {
			return err
		}

		// Update remaining fees
		remaining = remaining.Sub(validatorFeeShare)

		// Emit events for transparency
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		sdkCtx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeRewards,
				sdk.NewAttribute(sdk.AttributeKeyAmount, reward.String()),
				sdk.NewAttribute(types.AttributeKeyValidator, validator.GetOperator()),
				sdk.NewAttribute("fee_share", validatorFeeShare.String()),
				sdk.NewAttribute("emission", emission.String()),
			),
		)
	}

	// Apply community tax to remaining fees
	communityPoolShare := remaining.MulDecTruncate(communityTax)
	feePool.CommunityPool = feePool.CommunityPool.Add(communityPoolShare...)
	remaining = remaining.Sub(communityPoolShare)

	// Add any unallocated fees to community pool
	feePool.CommunityPool = feePool.CommunityPool.Add(remaining...)

	// Update fee pool
	err = k.FeePool.Set(ctx, feePool)
	if err != nil {
		return err
	}

	// Emit event for total emission
	if !totalEmission.IsZero() {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		sdkCtx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeWithdrawRewards,
				sdk.NewAttribute(sdk.AttributeKeyAmount, totalEmission.String()),
				sdk.NewAttribute("source", "validator_rewards_emission"),
			),
		)
	}

	return nil
}

// AllocateTokensToValidator allocates tokens to a validator, splitting according to commission.
func (k WrappedBaseKeeper) AllocateTokensToValidator(ctx context.Context, val stakingtypes.ValidatorI, tokens sdk.DecCoins) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	logger := k.Logger(sdkCtx)

	commissionRate := sdkmath.LegacyNewDecWithPrec(5, 2)
	commission := tokens.MulDec(commissionRate)

	valBz, err := k.sk.ValidatorAddressCodec().StringToBytes(val.GetOperator())
	if err != nil {
		return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[AllocateTokensToValidator][ValidatorAddressCodec] failed. failed to create codec for giving validator.")
	}

	valStr, err := k.sk.ValidatorAddressCodec().BytesToString(valBz)
	if err != nil {
		return err
	}

	// Update current commission
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCommission,
			sdk.NewAttribute(sdk.AttributeKeyAmount, commission.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, valStr),
		),
	)

	currentCommission, err := k.GetValidatorAccumulatedCommission(ctx, valBz)
	if err != nil {
		return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[AllocateTokensToValidator][GetValidatorAccumulatedCommission] failed. couldn't fetch validator commission.")
	}

	emission, found := k.apk.GetTotalEmission(ctx)
	if !found {
		return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[AllocateTokensToValidator][GetTotalEmission] failed. Total emission not found.")
	}

	decEmission, err := sdkmath.LegacyNewDecFromStr(emission.TotalEmission)
	if err != nil {
		return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[AllocateTokensToValidator][LegacyNewDecFromStr] failed. Total emission not converted.")
	}

	if decEmission.Sub(tokens[0].Amount).LT(tokens[0].Amount) {
		logger.Warn("[AllocateTokensToValidator] Failed to allocate reward to validators. Total emission is less than validator's reward.",
			"reward token", tokens[0].Amount,
			"total emission", emission.TotalEmission,
		)
		return nil
	}

	emission.TotalEmission = decEmission.Sub(tokens[0].Amount).String()
	err = k.apk.SetEmission(ctx, emission)
	if err != nil {
		return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[AllocateTokensToValidator][SetEmission] failed. Total emission not decreased.")
	}

	currentCommission.Commission = currentCommission.Commission.Add(commission...)
	err = k.SetValidatorAccumulatedCommission(ctx, valBz, currentCommission)
	if err != nil {
		return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[AllocateTokensToValidator][SetValidatorAccumulatedCommission] failed.")
	}

	currentRewards, err := k.GetValidatorCurrentRewards(ctx, valBz)
	if err != nil {
		return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[AllocateTokensToValidator][GetValidatorCurrentRewards] failed. couldn't fetch validator by giving validator address.")
	}

	currentRewards.Rewards = currentRewards.Rewards.Add(tokens...)
	err = k.SetValidatorCurrentRewards(ctx, valBz, currentRewards)
	if err != nil {
		return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[AllocateTokensToValidator][SetValidatorCurrentRewards] failed. couldn't set validator by giving validator address.")
	}
	logger.Warn("allocation reward is done for validator.", val.GetOperator(),
		"reward token", tokens[0].Amount,
		"total emission", emission.TotalEmission,
	)

	// Update outstanding rewards
	outstanding, err := k.GetValidatorOutstandingRewards(ctx, valBz)
	if err != nil {
		return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[AllocateTokensToValidator][GetValidatorOutstandingRewards] failed. couldn't fetch validator outstanding reward.")
	}
	outstanding.Rewards = outstanding.Rewards.Add(tokens...)

	return k.SetValidatorOutstandingRewards(ctx, valBz, outstanding)
}
