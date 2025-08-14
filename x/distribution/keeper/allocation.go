package keeper

import (
	"context"
	"strconv"

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

func (k WrappedBaseKeeper) MintCoins(ctx context.Context, newCoins sdk.Coins) error {
	if newCoins.Empty() {
		// skip as no coins need to be minted
		return nil
	}

	return k.bk.MintCoins(ctx, types.ModuleName, newCoins)
}

func (k WrappedBaseKeeper) CalculateBlockProvisioningReward(ctx context.Context) (mintedAmount sdkmath.Int, err error) {

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	logger := k.Logger(sdkCtx)

	minted, found := k.mk.GetTotalMinted(ctx)
	if !found {
		return sdkmath.Int{}, sdkerrors.Wrapf(errortypes.ErrInvalidRequest, "[CalculateBlockProvisioningReward] failed to retrieve total minted for block %d: no previous block provisioning reward found", sdkCtx.BlockHeight())
	}

	blockMinted := minted.BlockMinted.TruncateInt()
	if !blockMinted.IsPositive() {
		return sdkmath.Int{}, sdkerrors.Wrapf(errortypes.ErrInvalidCoins, "[CalculateBlockProvisioningReward] failed to calculate block provisioning reward: block minted amount must be positive, got %s", blockMinted.String())
	}

	// blockMinted := sdkmath.NewIntFromUint64(12)

	mintedCoin := sdk.NewCoin(params.BondDenom, blockMinted)
	mintedCoins := sdk.NewCoins(mintedCoin)
	err = k.MintCoins(ctx, mintedCoins)
	if err != nil {
		return sdkmath.Int{}, sdkerrors.Wrapf(errortypes.ErrInvalidCoins, "[CalculateBlockProvisioningReward] failed to mint block provisioning reward for block %d: invalid coin amount %s", sdkCtx.BlockHeight(), mintedCoin.String())
	}

	emission, found := k.apk.GetEmissionEntityPerBlock(ctx, strconv.FormatInt(sdkCtx.BlockHeight(), 10))
	if !found {
		return sdkmath.Int{}, sdkerrors.Wrapf(errortypes.ErrInvalidRequest, "[CalculateBlockProvisioningReward] failed to retrieve total emission for block %d: no previous emission data found", sdk.UnwrapSDKContext(ctx).BlockHeight())
	}

	decEmission, err := sdkmath.LegacyNewDecFromStr(emission.RemainBlockEmission)
	if err != nil {
		return sdkmath.Int{}, sdkerrors.Wrapf(errortypes.ErrInvalidRequest, "[CalculateBlockProvisioningReward] failed to create decimal for block provisioning reward in block %d: invalid input string %s", sdk.UnwrapSDKContext(ctx).BlockHeight(), decEmission.String())
	}

	if decEmission.Sub(minted.BlockMinted).LT(minted.BlockMinted) {
		logger.Warn("[CalculateBlockProvisioningReward] Skipping validator due to insufficient emission tokens.",
			"total_mintes", minted.BlockMinted,
		)
		return sdkmath.Int{}, nil
	}

	if decEmission.Sub(minted.BlockMinted).IsNegative() || decEmission.Sub(minted.BlockMinted).IsZero() {
		logger.Warn("[CalculateBlockProvisioningReward] Validator reward exceeds available emission.",
			"total_mintedn", minted.BlockMinted,
		)
		return sdkmath.Int{}, nil
	}

	emission.RemainBlockEmission = decEmission.Sub(minted.BlockMinted).String()
	err = k.apk.SetEmission(ctx, emission)
	if err != nil {
		logger.Error("[CalculateBlockProvisioningReward] Failed to update emission pool after deducting validator reward.",
			"error", err.Error(),
			"remaining_emission", decEmission.String(),
		)
		return sdkmath.Int{}, sdkerrors.Wrapf(errortypes.ErrInvalidRequest,
			"[CalculateBlockProvisioningReward] failed to update emission pool after deducting validator reward: %v", err)
	}

	return blockMinted, nil
}

// AllocateTokens performs reward and fee distribution to all validators based
// on the F1 fee distribution specification.
func (k WrappedBaseKeeper) AllocateTokens(ctx context.Context, totalPreviousPower int64, validators []stakingtypes.Validator) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	logger := k.Logger(sdkCtx)

	logger.Info("[💸 Distribution] Starting initiating validator token allocation for block")

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
		"fees collected", feesCollected.String(),
		"block_height", sdkCtx.BlockHeight(),
	)

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
			logger.Info("## Fee extracted from the fee_collector",
				"feesCollected", feesCollected,
			)
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

		logger.Info("#### Calculation of transaction reward per validator",
			"validator", validator.GetOperator(),
			"validator fee share", validatorFeeShare,
			"min apr", minAprRewardAmount,
			"need mint", validatorFeeShare[0].Amount.LT(minAprReward[0].Amount),
			"must mint amount", (minAprReward[0].Amount.Sub(validatorFeeShare[0].Amount)).String(),
		)

		var reward sdk.DecCoins

		if validatorFeeShare[0].Amount.GTE(minAprReward[0].Amount) {
			logger.Info("## Fee share is sufficient, no need to mint reward",
				"reward", validatorFeeShare[0].Amount,
				"min apr reward", minAprReward[0].Amount,
			)
			reward = validatorFeeShare
		} else {
			logger.Info("## Fee share is insufficient, mint for minimum APR reward",
				"reward", validatorFeeShare[0].Amount,
				"min apr reward", minAprReward[0].Amount,
			)
			// Fee share is insufficient, use minimum APR reward
			mustEmission := minAprReward[0].Amount.Sub(validatorFeeShare[0].Amount)
			if mustEmission.IsNegative() || mustEmission.IsZero() || mustEmission.IsNil() {
				logger.Warn("Mint amount is not valid.")
				return sdkerrors.Wrap(errortypes.ErrInvalidType, "[AllocateTokens] failed to deduct validator fee share from min apr amount")
			}

			err = k.bk.MintCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(params.BondDenom, mustEmission.TruncateInt())))
			if err != nil {
				return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[AllocateTokens][MintCoins] failed. couldn't mint coin.")
			}

			reward = sdk.NewDecCoins(sdk.NewDecCoinFromDec(params.BondDenom, minAprReward[0].Amount))

			emission, found := k.apk.GetEmissionEntityPerBlock(ctx, strconv.FormatInt(sdkCtx.BlockHeight(), 10))
			if !found {
				return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[AllocateTokens][GetTotalEmission] failed. Total emission not found.")
			}

			decEmission, err := sdkmath.LegacyNewDecFromStr(emission.RemainBlockEmission)
			if err != nil {
				return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[AllocateTokens][LegacyNewDecFromStr] failed. couldn't create a decimal from an input decimal string.")
			}

			if decEmission.Sub(mustEmission).LT(mustEmission) {
				logger.Warn("[AllocateTokensToValidator] Skipping validator due to insufficient emission tokens.",
					"validator", validator.GetOperator(),
					"required_emission", mustEmission.String(),
					"total_emission", emission.RemainBlockEmission,
				)
				return nil
			}

			if decEmission.Sub(mustEmission).IsNegative() || decEmission.Sub(mustEmission).IsZero() {
				logger.Warn("[AllocateTokens] skipping validator due to invalid emission tokens.",
					"total_emission", decEmission.String(),
					"required_emission", mustEmission.String(),
					"validator", validator.GetOperator(),
				)
				return sdkerrors.Wrap(errortypes.ErrInvalidType, "[AllocateTokens] insufficient emission to cover validator reward requiered emission")
			}

			emission.RemainBlockEmission = decEmission.Sub(mustEmission).String()
			err = k.apk.SetEmission(ctx, emission)
			if err != nil {
				logger.Error("[AllocateTokens] failed to update emission pool after deducting validator reward.",
					"error", err.Error(),
					"remaining_emission", decEmission.String(),
				)
				return sdkerrors.Wrapf(errortypes.ErrInvalidRequest,
					"failed to update emission pool after deducting validator reward: %v", err)
			}

			logger.Info("[AllocateTokens] total emission in the allocation pool successfully updated.",
				"block_height", sdkCtx.BlockHeight(),
				"emission_amount", mustEmission.String(),
			)
		}

		err = k.AllocateTokensToValidator(ctx, validator, reward)
		if err != nil {
			logger.Error("[AllocateTokens] failed to allocate tokens to validator.",
				"validator", validator.GetOperator(),
				"reward", reward[0].Amount.String(),
				"error", err.Error(),
			)
			return sdkerrors.Wrapf(errortypes.ErrInvalidRequest,
				"[AllocateTokens][AllocateTokensToValidator] failed for validator %s: %v",
				validator.GetOperator(), err)
		}

		logger.Info("[AllocateTokens] successfully allocated tokens to validator.",
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

// AllocateTokens performs reward and fee distribution to all validators based
// on the F1 fee distribution specification.
func (k WrappedBaseKeeper) AllocateBlockProvisioningTokens(ctx context.Context, totalPreviousPower int64, bondedVotes []abci.VoteInfo) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	logger := k.Logger(sdkCtx)

	rewardAmount, _ := k.CalculateBlockProvisioningReward(ctx)

	if rewardAmount.IsNil() || rewardAmount.IsNegative() || rewardAmount.IsZero() {
		logger.Warn("[AllocateBlockProvisioningTokens] Skipping validator due to insufficient emission tokens.",
			"reward_amount", rewardAmount,
		)
		return nil
	}

	feesCollected := sdk.NewDecCoinsFromCoins(sdk.NewCoin(params.BondDenom, rewardAmount))

	// temporary workaround to keep CanWithdrawInvariant happy
	// general discussions here: https://github.com/cosmos/cosmos-sdk/issues/2906#issuecomment-441867634
	feePool, err := k.FeePool.Get(ctx)
	if err != nil {
		return err
	}

	if totalPreviousPower == 0 {
		feePool.CommunityPool = feePool.CommunityPool.Add(feesCollected...)
		return k.FeePool.Set(ctx, feePool)
	}

	// calculate fraction allocated to validators
	remaining := feesCollected
	communityTax, err := k.GetCommunityTax(ctx)
	if err != nil {
		return err
	}

	var previousTotalPower int64
	for _, voteInfo := range sdkCtx.VoteInfos() {
		previousTotalPower += voteInfo.Validator.Power
	}

	voteMultiplier := sdkmath.LegacyOneDec().Sub(communityTax)
	feeMultiplier := feesCollected.MulDecTruncate(voteMultiplier)

	for _, vote := range bondedVotes {
		validator, err := k.sk.ValidatorByConsAddr(ctx, vote.Validator.Address)
		if err != nil {
			return err
		}

		// TODO: Consider micro-slashing for missing votes.
		//
		// Ref: https://github.com/cosmos/cosmos-sdk/issues/2525#issuecomment-430838701
		powerFraction := sdkmath.LegacyNewDec(vote.Validator.Power).QuoTruncate(sdkmath.LegacyNewDec(totalPreviousPower))
		reward := feeMultiplier.MulDecTruncate(powerFraction)

		err = k.AllocateBlockProvisioningTokensToValidator(ctx, validator, reward)
		if err != nil {
			return err
		}

		remaining = remaining.Sub(reward)
	}

	// allocate community funding
	feePool.CommunityPool = feePool.CommunityPool.Add(remaining...)
	return k.FeePool.Set(ctx, feePool)
}

// AllocateTokensToValidator allocate tokens to a particular validator,
// splitting according to commission.
func (k WrappedBaseKeeper) AllocateBlockProvisioningTokensToValidator(ctx context.Context, val stakingtypes.ValidatorI, tokens sdk.DecCoins) error {
	// split tokens between validator and delegators according to commission
	commission := tokens.MulDec(val.GetCommission())
	shared := tokens.Sub(commission)

	valBz, err := k.sk.ValidatorAddressCodec().StringToBytes(val.GetOperator())
	if err != nil {
		return err
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
		return err
	}

	currentCommission.Commission = currentCommission.Commission.Add(commission...)
	err = k.SetValidatorAccumulatedCommission(ctx, valBz, currentCommission)
	if err != nil {
		return err
	}

	// update current rewards
	currentRewards, err := k.GetValidatorCurrentRewards(ctx, valBz)
	if err != nil {
		return err
	}

	currentRewards.Rewards = currentRewards.Rewards.Add(shared...)
	err = k.SetValidatorCurrentRewards(ctx, valBz, currentRewards)
	if err != nil {
		return err
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
		return err
	}

	outstanding.Rewards = outstanding.Rewards.Add(tokens...)
	return k.SetValidatorOutstandingRewards(ctx, valBz, outstanding)
}

var (
	// DefaultPowerReduction is the default amount of staking tokens required for 1 unit of consensus-engine power
	DefaultPowerReduction = sdkmath.NewIntFromUint64(1000000)
)

func (k WrappedBaseKeeper) PowerReduction(ctx context.Context) sdkmath.Int {
	return DefaultPowerReduction
}

func (k WrappedBaseKeeper) TokensToConsensusPower(ctx context.Context, tokens sdkmath.Int) int64 {
	return TokensToConsensusPower(tokens, k.PowerReduction(ctx))
}

func TokensToConsensusPower(tokens, powerReduction sdkmath.Int) int64 {
	return (tokens.Quo(powerReduction)).Int64()
}
