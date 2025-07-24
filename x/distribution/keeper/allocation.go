package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/outbe/outbe-node/app/params"
	errortypes "github.com/outbe/outbe-node/errors"
	"github.com/outbe/outbe-node/x/distribution/constants"
)

// AllocateTokens performs reward and fee distribution to all validators based
// on the F1 fee distribution specification.
func (k WrappedBaseKeeper) AllocateTokens(ctx context.Context, totalPreviousPower int64, bondedVotes []abci.VoteInfo) error {

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	logger := k.Logger(sdkCtx)

	feeCollector := k.ak.GetModuleAccount(ctx, constants.FeeCollectorName)
	feesCollectedInt := k.bk.GetAllBalances(ctx, feeCollector.GetAddress())
	feesCollected := sdk.NewDecCoinsFromCoins(feesCollectedInt...)

	if feesCollected.IsZero() || !feesCollected.IsValid() || feesCollected.Empty() {
		return nil
	}

	// transfer collected fees to the distribution module account
	err := k.bk.SendCoinsFromModuleToModule(ctx, constants.FeeCollectorName, types.ModuleName, feesCollectedInt)
	if err != nil {
		return sdkerrors.Wrapf(errortypes.ErrInvalidRequest, "[AllocateTokensToValidator][SendCoinsFromModuleToModule] failed. couldn't send coin fron fee collector to distribution module. %s", err)
	}

	/// Get fee pool for community pool updates
	feePool, err := k.FeePool.Get(ctx)
	if err != nil {
		return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[AllocateTokens][FeePool.Get] failed. couldn't fetch fee pool.")
	}

	// Handle case with no validators
	if len(bondedVotes) == 0 {
		feePool.CommunityPool = feePool.CommunityPool.Add(feesCollected...)
		return k.FeePool.Set(ctx, feePool)
	}

	// Calculate total self-bonded tokens across all validators
	totalSelfBonded, err := k.rk.CalculateTotalSelfBondedTokens(ctx, bondedVotes)
	if err != nil {
		return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[AllocateTokens][CalculateTotalSelfBondedTokens] failed. couldn't calculate total self bond token.")
	}

	if totalSelfBonded.IsZero() || totalSelfBonded.IsNil() {
		feePool.CommunityPool = feePool.CommunityPool.Add(feesCollected...)
		return k.FeePool.Set(ctx, feePool)
	}

	// Initialize remaining fees and total emission
	remaining := feesCollected

	validators, err := k.sk.GetAllValidators(ctx)
	if err != nil {
		return sdkerrors.Wrap(errortypes.ErrInvalidType, "failed to query validator.")
	}

	for _, validator := range validators {
		conAddress, _ := validator.GetConsAddr()
		validator, err := k.sk.ValidatorByConsAddr(ctx, conAddress)
		if err != nil {
			return sdkerrors.Wrapf(errortypes.ErrInvalidRequest, "[AllocateTokensToValidator][AllocateTokensToValidator] failed. couldn't fetch validator cons codec. %s", err)
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
			if mustMint.IsNegative() || mustMint.IsZero() {
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
			reward = sdk.NewDecCoins(sdk.NewDecCoinFromDec(validatorFeeShare[0].Denom, newAmount))
		}

		emission, found := k.apk.GetTotalEmission(ctx)
		if !found {
			return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[AllocateTokens][GetTotalEmission] failed. Total emission not found.")
		}

		decEmission, _ := math.LegacyNewDecFromStr(emission.TotalEmission)
		if decEmission.Sub(reward[0].Amount).LT(reward[0].Amount) {
			logger.Warn("[AllocateTokensToValidator] Failed to allocate reward to validators. Total emission is less than validator's reward.",
				"reward token", reward[0].Amount,
				"total emission", emission.TotalEmission,
			)
			return nil
		}
		if decEmission.Sub(reward[0].Amount).IsNegative() || decEmission.Sub(reward[0].Amount).IsZero() {
			logger.Warn("Deduction emission from rewardt is not valid.",
				"mustMint", decEmission,
				"reward", reward,
			)
			return sdkerrors.Wrap(errortypes.ErrInvalidType, "[AllocateTokens] failed to deduct validator fee share from min apr amount")
		}
		emission.TotalEmission = decEmission.Sub(reward[0].Amount).String()
		err = k.apk.SetEmission(ctx, emission)
		if err != nil {
			return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[AllocateTokens][SetEmission] failed. Total emission not decreased.")
		}

		err = k.AllocateTokensToValidator(ctx, validator, reward)
		if err != nil {
			return sdkerrors.Wrapf(errortypes.ErrInvalidRequest, "[AllocateTokensToValidator][AllocateTokensToValidator] failed. couldn't create validator address codec. %s", err)
		}

		remaining = remaining.Sub(validatorFeeShare)

	}

	// allocate community funding
	feePool.CommunityPool = feePool.CommunityPool.Add(remaining...)
	return k.FeePool.Set(ctx, feePool)
}

// AllocateTokensToValidator allocate tokens to a particular validator,
// splitting according to commission.
func (k WrappedBaseKeeper) AllocateTokensToValidator(ctx context.Context, val stakingtypes.ValidatorI, tokens sdk.DecCoins) error {
	commission := tokens.MulDec(val.GetCommission())

	valBz, err := k.sk.ValidatorAddressCodec().StringToBytes(val.GetOperator())
	if err != nil {
		return sdkerrors.Wrapf(errortypes.ErrInvalidRequest, "[AllocateTokensToValidator][ValidatorAddressCodec] failed. couldn't create validator address codec. %s", err)
	}

	// update current commission
	sdkCtx := sdk.UnwrapSDKContext(ctx)
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
	return k.SetValidatorOutstandingRewards(ctx, valBz, outstanding)
}
