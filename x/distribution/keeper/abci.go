package keeper

import (
	"context"
	"strconv"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/outbe/outbe-node/app/params"
	errortypes "github.com/outbe/outbe-node/errors"
)

// BeginBlocker will persist the current header and validator set as a historical entry
// and prune the oldest entry based on the HistoricalEntries parameter
func (k WrappedBaseKeeper) BeginBlocker(ctx context.Context) error {
	defer telemetry.ModuleMeasureSince(stakingtypes.ModuleName, telemetry.Now(), telemetry.MetricKeyBeginBlocker)

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	logger := k.Logger(sdkCtx)

	wctx := sdk.UnwrapSDKContext(ctx)

	// determine the total power signing the block
	var previousTotalPower int64

	for _, voteInfo := range wctx.VoteInfos() {
		previousTotalPower += voteInfo.Validator.Power
	}

	if wctx.BlockHeight() > 1 {
		if err := k.AllocateBlockProvisioningTokens(ctx, previousTotalPower, wctx.VoteInfos()); err != nil {
			return sdkerrors.Wrapf(err, "[BeginBlocker][Distribution] failed to distribute block provisioning rewards for a validator at block %d", sdkCtx.BlockHeight())
		}
	}

	validators, err := k.sk.GetAllValidators(ctx)
	if err != nil {
		return sdkerrors.Wrapf(errortypes.ErrInvalidType,
			"[BeginBlocker][Distribution] failed to query all validators: %v", err)
	}

	if err := k.AllocateTokens(ctx, previousTotalPower, validators); err != nil {
		logger.Error("[BeginBlocker][Distribution] failed to distribute validator rewards", "error", err.Error())
		return sdkerrors.Wrapf(err, "[BeginBlocker][Distribution] failed to distribute transactions fee for a validator at block %d", sdkCtx.BlockHeight())
	}

	emission, found := k.apk.GetEmissionEntityPerBlock(ctx, strconv.FormatInt(sdkCtx.BlockHeight(), 10))
	if !found {
		return sdkerrors.Wrap(err, "[GetEmissionPerBlock] failed to fetch emission per block")
	}

	emissionPerblockDec, err := sdkmath.LegacyNewDecFromStr(emission.RemainBlockEmission)
	if err != nil {
		return sdkerrors.Wrap(err, "failed to pars emission amount from string to dec.")
	}

	if emissionPerblockDec.IsNegative() || emissionPerblockDec.IsNil() {
		return sdkerrors.Wrap(err, "no emission left for CRA reward.")
	}

	// 8% of token in current block limit for cra
	craRewardPerBlock := emissionPerblockDec.Mul(sdkmath.LegacyNewDecWithPrec(8, 2))

	if emissionPerblockDec.Sub(craRewardPerBlock).LT(craRewardPerBlock) {
		logger.Warn("Skipping cra reward due to insufficient emission tokens.",
			"cra_reward_per_block", craRewardPerBlock,
		)
		return nil
	}

	if emissionPerblockDec.Sub(craRewardPerBlock).IsNegative() || emissionPerblockDec.Sub(craRewardPerBlock).IsZero() {
		logger.Warn(" cra reward exceeds available emission.",
			"cra_reward_per_block", craRewardPerBlock,
		)
		return nil
	}

	cras := k.apk.GetCRAAll(ctx)
	for _, cra := range cras {
		craReward := craRewardPerBlock.Quo(sdkmath.LegacyNewDec(2)).Mul(sdkmath.LegacyNewDecWithPrec(32, 2))

		cra, found := k.apk.GetCRAByCRAAddress(ctx, cra.CraAddress)
		if !found {
			return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "failed to fetch a valid cra for giving address")
		}
		cra.Reward.Add(craReward)
		k.apk.SetCRA(ctx, cra)

		newCoin := sdk.NewDecCoinFromDec(params.BondDenom, craReward)
		coin := newCoin.Amount.TruncateDec()
		k.bk.MintCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(params.BondDenom, coin.TruncateInt())))

		err := k.bk.SendCoinsFromModuleToModule(ctx, types.ModuleName, "cra", sdk.NewCoins(sdk.NewCoin(params.BondDenom, coin.TruncateInt())))
		if err != nil {
			logger.Error("Failed to transfer reward to cra module",
				"error", err,
				"block_height", sdkCtx.BlockHeight())
			return sdkerrors.Wrap(err, "failed to transfer reward to cra module")
		}
	}

	wallets := k.apk.GetWalletAll(ctx)
	for _, wallet := range wallets {
		walletReward := craRewardPerBlock.Quo(sdkmath.LegacyNewDec(2)).Mul(sdkmath.LegacyNewDecWithPrec(32, 2))

		wallet, found := k.apk.GetWalletByCRAAddress(ctx, wallet.Address)
		if !found {
			return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "failed to fetch a valid wallet for giving address")
		}
		wallet.Reward.Add(walletReward)
		k.apk.SetWallet(ctx, wallet)

		newCoin := sdk.NewDecCoinFromDec(params.BondDenom, walletReward)
		coin := newCoin.Amount.TruncateDec()
		k.bk.MintCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(params.BondDenom, coin.TruncateInt())))

		newCoin = sdk.NewDecCoinFromDec(params.BondDenom, walletReward)
		coin = newCoin.Amount.TruncateDec()
		k.bk.MintCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(params.BondDenom, coin.TruncateInt())))

		err := k.bk.SendCoinsFromModuleToModule(ctx, types.ModuleName, "cra", sdk.NewCoins(sdk.NewCoin(params.BondDenom, coin.TruncateInt())))
		if err != nil {
			logger.Error("Failed to transfer reward to cra module",
				"error", err,
				"block_height", sdkCtx.BlockHeight())
			return sdkerrors.Wrap(err, "failed to transfer reward to cra module")
		}
	}

	// record the proposer for when we payout on the next block
	consAddr := sdk.ConsAddress(wctx.BlockHeader().ProposerAddress)
	return k.SetPreviousProposerConsAddr(ctx, consAddr)
}
