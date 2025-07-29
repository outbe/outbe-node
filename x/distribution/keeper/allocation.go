package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/outbe/outbe-node/app/params"
	errortypes "github.com/outbe/outbe-node/errors"
	"github.com/outbe/outbe-node/x/distribution/constants"
)

// AllocateTokens performs reward and fee distribution to all validators based
// on the F1 fee distribution specification.
func (k WrappedBaseKeeper) AllocateTokens(ctx context.Context, totalPreviousPower int64, validators []stakingtypes.Validator) error {

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	logger := k.Logger(sdkCtx)

	logger.Info("################## Starting validator token allocation",
		"block_height", sdkCtx.BlockHeight(),
		"module", "distribution",
	)

	feeCollector := k.ak.GetModuleAccount(ctx, constants.FeeCollectorName)
	feesCollectedInt := k.bk.GetAllBalances(ctx, feeCollector.GetAddress())
	feesCollected := sdk.NewDecCoinsFromCoins(feesCollectedInt...)

	if feesCollected.IsZero() || !feesCollected.IsValid() {
		logger.Info("Skipping token allocation: no valid fees collected",
			"block_height", sdkCtx.BlockHeight(),
			"fee_collected", feesCollectedInt)
		return nil
	}

	// transfer collected fees to the distribution module account
	err := k.bk.SendCoinsFromModuleToModule(ctx, constants.FeeCollectorName, types.ModuleName, feesCollectedInt)
	if err != nil {
		logger.Error("Failed to transfer fees from fee collector to distribution module",
			"error", err,
			"block_height", sdkCtx.BlockHeight())
		return sdkerrors.Wrap(err, "failed to transfer fees to distribution module")
	}

	logger.Info("Successfully transferred fees to distribution module",
		"module", "distribution",
		"block_height", sdkCtx.BlockHeight())

	// Get fee pool for community pool updates
	feePool, err := k.FeePool.Get(ctx)
	if err != nil {
		return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[AllocateTokens][FeePool.Get] failed. couldn't fetch fee pool.")
	}

	// Handle case with no validators
	if len(validators) == 0 {
		feePool.CommunityPool = feePool.CommunityPool.Add(feesCollected...)
		return k.FeePool.Set(ctx, feePool)
	}

	// Calculate total self-bonded tokens across all validators
	totalSelfBonded, err := k.rk.CalculateTotalSelfBondedTokens(ctx, validators)
	if err != nil {
		return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[AllocateTokens][CalculateTotalSelfBondedTokens] failed. couldn't calculate total self bond token.")
	}

	if totalSelfBonded.IsZero() || totalSelfBonded.IsNil() {
		feePool.CommunityPool = feePool.CommunityPool.Add(feesCollected...)
		return k.FeePool.Set(ctx, feePool)
	}

	// Initialize remaining fees and total emission
	remaining := feesCollected

	for _, validator := range validators {
		conAddress, err := validator.GetConsAddr()
		if err != nil {
			return sdkerrors.Wrapf(errortypes.ErrInvalidRequest, "[AllocateTokens][GetConsAddr] failed. couldn't extracts Consensus key address. %s", err)
		}

		validator, err := k.sk.ValidatorByConsAddr(ctx, conAddress)
		if err != nil {
			return sdkerrors.Wrapf(errortypes.ErrInvalidRequest, "[AllocateTokens][AllocateTokensToValidator] failed. couldn't fetch validator cons codec. %s", err)
		}

		if !validator.IsBonded() || validator.IsJailed() || validator.IsUnbonded() || validator.IsUnbonded() {
			return sdkerrors.Wrapf(errortypes.ErrInvalidRequest, "[AllocateTokens] failed. validator is not bonded validator. Couldn't get reward")
		}

		// Get validator's self-bonded tokens
		selfBonded, err := k.rk.GetValidatorSelfBondedTokens(ctx, validator)
		if err != nil {
			return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[AllocateTokens][GetValidatorSelfBondedTokens] failed. couldn't fetch validator self bond token.")
		}

		var validatorFeeShare sdk.DecCoins
		for _, fee := range feesCollected {
			if fee.Denom == params.BondDenom {
				share := k.rk.CalculateFeeShare(fee.Amount, selfBonded, totalSelfBonded)
				validatorFeeShare = validatorFeeShare.Add(sdk.NewDecCoinFromDec(fee.Denom, share))
			}
		}

		minAprRewardAmount, err := k.rk.CalculateMinApr(ctx, selfBonded)
		if err != nil {
			return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[AllocateTokens][CalculateMinApr] failed. couldn't calculate min apr.")
		}

		minAprReward := sdk.NewDecCoins(sdk.NewDecCoinFromDec(params.BondDenom, minAprRewardAmount))

		var reward sdk.DecCoins

		if validatorFeeShare[0].Amount.GT(minAprReward[0].Amount) {
			reward = validatorFeeShare
		} else {
			// Fee share is insufficient, use minimum APR reward
			mustMint := minAprReward[0].Amount.Sub(validatorFeeShare[0].Amount)
			if mustMint.IsNegative() || mustMint.IsZero() || mustMint.IsNil() {
				logger.Warn("Mint amount is not valid.",
					"mustMint", mustMint,
				)
				return sdkerrors.Wrap(errortypes.ErrInvalidType, "[AllocateTokens] failed to deduct validator fee share from min apr amount")
			}

			err = k.bk.MintCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(params.BondDenom, mustMint.TruncateInt())))
			if err != nil {
				return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[AllocateTokens][MintCoins] failed. couldn't mint coin.")
			}

			newAmount := validatorFeeShare[0].Amount.Add(mustMint)
			reward = sdk.NewDecCoins(sdk.NewDecCoinFromDec(params.BondDenom, newAmount))
		}

		emission, found := k.apk.GetTotalEmission(ctx)
		if !found {
			return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[AllocateTokens][GetTotalEmission] failed. Total emission not found.")
		}

		decEmission, err := math.LegacyNewDecFromStr(emission.TotalEmission)
		if err != nil {
			return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[AllocateTokens][LegacyNewDecFromStr] failed. couldn't create a decimal from an input decimal string.")
		}

		if decEmission.Sub(reward[0].Amount).LT(reward[0].Amount) {
			logger.Warn("[AllocateTokensToValidator] Skipping validator due to insufficient emission tokens.",
				"validator", validator.GetOperator(),
				"required_reward", reward[0].Amount.String(),
				"total_emission", emission.TotalEmission,
			)
			return nil
		}

		if decEmission.Sub(reward[0].Amount).IsNegative() || decEmission.Sub(reward[0].Amount).IsZero() {
			logger.Warn("[AllocateTokens] Validator reward exceeds available emission.",
				"total_emission", decEmission.String(),
				"required_reward", reward[0].Amount.String(),
				"validator", validator.GetOperator(),
			)
			return sdkerrors.Wrap(errortypes.ErrInvalidType, "[AllocateTokens] insufficient emission to cover validator reward")
		}

		emission.TotalEmission = decEmission.Sub(reward[0].Amount).String()
		err = k.apk.SetEmission(ctx, emission)
		if err != nil {
			logger.Error("[AllocateTokens] Failed to update emission pool after deducting validator reward.",
				"error", err.Error(),
				"remaining_emission", decEmission.String(),
			)
			return sdkerrors.Wrapf(errortypes.ErrInvalidRequest,
				"failed to update emission pool after deducting validator reward: %v", err)
		}

		logger.Info("[AllocateTokens] Total emission in the allocation pool successfully updated.",
			"block_height", sdkCtx.BlockHeight(),
			"reward_amount", reward[0].Amount.String(),
		)

		err = k.AllocateTokensToValidator(ctx, validator, reward)
		if err != nil {
			logger.Error("[AllocateTokens] Failed to allocate tokens to validator.",
				"validator", validator.GetOperator(),
				"reward", reward[0].Amount.String(),
				"error", err.Error(),
			)
			return sdkerrors.Wrapf(errortypes.ErrInvalidRequest,
				"[AllocateTokens][AllocateTokensToValidator] failed for validator %s: %v",
				validator.GetOperator(), err)
		}

		logger.Info("[AllocateTokens] Successfully allocated tokens to validator.",
			"block_height", sdkCtx.BlockHeight(),
			"validator", validator.GetOperator(),
			"reward", reward.String(),
		)

		remaining = remaining.Sub(validatorFeeShare)
	}

	// allocate community funding
	feePool.CommunityPool = feePool.CommunityPool.Add(remaining...)
	return k.FeePool.Set(ctx, feePool)
}

// AllocateTokensToValidator allocate tokens to a particular validator,
// splitting according to commission.
func (k WrappedBaseKeeper) AllocateTokensToValidator(ctx context.Context, val stakingtypes.ValidatorI, tokens sdk.DecCoins) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	logger := k.Logger(sdkCtx)

	commission := tokens.MulDec(val.GetCommission())

	valBz, err := k.sk.ValidatorAddressCodec().StringToBytes(val.GetOperator())
	if err != nil {
		logger.Error("[AllocateTokensToValidator] Failed to convert validator address to bytes.",
			"validator", val.GetOperator(),
			"error", err.Error(),
		)
		return sdkerrors.Wrapf(errortypes.ErrInvalidRequest,
			"[AllocateTokensToValidator] could not convert validator address %s to bytes: %v",
			val.GetOperator(), err)
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCommission,
			sdk.NewAttribute(sdk.AttributeKeyAmount, commission.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, val.GetOperator()),
		),
	)

	currentCommission, err := k.GetValidatorAccumulatedCommission(ctx, valBz)
	if err != nil {
		return sdkerrors.Wrapf(errortypes.ErrInvalidRequest, "[AllocateTokensToValidator][GetValidatorAccumulatedCommission] failed. couldn't fetch validator comission. %s", err)
	}

	currentCommission.Commission = currentCommission.Commission.Add(commission...)
	err = k.SetValidatorAccumulatedCommission(ctx, valBz, currentCommission)
	if err != nil {
		return sdkerrors.Wrapf(errortypes.ErrInvalidRequest, "[AllocateTokensToValidator][SetValidatorAccumulatedCommission] failed. couldn't set validator comission. %s", err)
	}

	// update current rewards
	currentRewards, err := k.GetValidatorCurrentRewards(ctx, valBz)
	if err != nil {
		return sdkerrors.Wrapf(errortypes.ErrInvalidRequest, "[AllocateTokensToValidator][GetValidatorCurrentRewards] failed. couldn't fetch validator current reward. %s", err)
	}

	currentRewards.Rewards = currentRewards.Rewards.Add(tokens...)
	err = k.SetValidatorCurrentRewards(ctx, valBz, currentRewards)
	if err != nil {
		return sdkerrors.Wrapf(errortypes.ErrInvalidRequest, "[AllocateTokensToValidator][SetValidatorCurrentRewards] failed. couldn't set validator current reward. %s", err)
	}

	// update outstanding rewards
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRewards,
			sdk.NewAttribute(sdk.AttributeKeyAmount, tokens.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, val.GetOperator()),
		),
	)

	outstanding, err := k.GetValidatorOutstandingRewards(ctx, valBz)
	if err != nil {
		return sdkerrors.Wrapf(errortypes.ErrInvalidRequest, "[AllocateTokensToValidator][GetValidatorOutstandingRewards] failed. couldn't fetch validator outstanding reward. %s", err)
	}

	outstanding.Rewards = outstanding.Rewards.Add(tokens...)

	logger.Info("################## Token allocation to validators completed successfully.",
		"block_height", sdkCtx.BlockHeight(),
	)

	return k.SetValidatorOutstandingRewards(ctx, valBz, outstanding)
}
