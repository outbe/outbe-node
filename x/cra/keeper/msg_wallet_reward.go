package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/outbe/outbe-node/app/params"
	errortypes "github.com/outbe/outbe-node/errors"
	"github.com/outbe/outbe-node/x/cra/constants"
	"github.com/outbe/outbe-node/x/cra/types"
)

func (k msgServer) WalletReward(goCtx context.Context, msg *types.MsgWalletReward) (*types.MsgWalletRewardResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	logger := k.Logger(ctx)

	logger.Info("🔁 Starting wallet reward transaction")

	if msg.Address == "" {
		return &types.MsgWalletRewardResponse{}, sdkerrors.Wrap(errortypes.ErrInvalidAddress, "[WalletReward] cra address can not be empty.")
	}

	if msg.Creator == "" {
		return &types.MsgWalletRewardResponse{}, sdkerrors.Wrap(errortypes.ErrInvalidAddress, "[WalletReward] creator can not be empty.")
	}

	wallet, found := k.GetWalletByWalletAddress(ctx, msg.Address)
	if !found {
		return &types.MsgWalletRewardResponse{}, sdkerrors.Wrap(errortypes.ErrInvalidAddress, "[WalletReward][GetCRAByCRAAddress] failed. couldn't fetch a valid cra.")
	}

	spendableCoin := sdk.NewCoins(sdk.NewCoin(params.BondDenom, wallet.Reward.TruncateInt().Sub(sdkmath.NewInt(constants.RewardDefragment))))

	err := k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		sdk.AccAddress(wallet.Address),
		spendableCoin,
	)
	if err != nil {
		return &types.MsgWalletRewardResponse{}, sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[WalletReward][SendCoinsFromModuleToAccount] failed. couldn't send coin to cra address.")
	}

	newWallet := types.Wallet{
		Creator: wallet.Creator,
		Address: wallet.Address,
		Reward:  sdkmath.LegacyNewDec(constants.RewardDefragment),
	}
	k.SetWallet(ctx, newWallet)

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventType,
			sdk.NewAttribute(types.AttributeCraAddress, wallet.Address),
		),
	)

	logger.Info("✅ Submitting a wallet reward transaction successfully completed")

	return &types.MsgWalletRewardResponse{}, nil
}
