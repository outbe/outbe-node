package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/outbe/outbe-node/app/params"
	errortypes "github.com/outbe/outbe-node/errors"
	"github.com/outbe/outbe-node/x/allocationpool/types"
)

func (k msgServer) WalletReward(goCtx context.Context, msg *types.MsgWalletReward) (*types.MsgWalletRewardResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	logger := k.Logger(ctx)

	logger.Info("🔁 Starting wallet reward transaction")

	if msg.Address == "" {
		return &types.MsgWalletRewardResponse{}, sdkerrors.Wrap(errortypes.ErrInvalidAddress, "cra address can not be empty.")
	}

	wallet, found := k.GetWalletByCRAAddress(ctx, msg.Address)
	if !found {
		return &types.MsgWalletRewardResponse{}, sdkerrors.Wrap(errortypes.ErrInvalidAddress, "[WalletReward][GetCRAByCRAAddress] failed. couldn't fetch a valid cra.")
	}

	err := k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		sdk.AccAddress(wallet.Address),
		sdk.NewCoins(sdk.NewCoin(params.BondDenom, sdkmath.Int(wallet.Reward))),
	)
	if err != nil {
		return &types.MsgWalletRewardResponse{}, sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[WalletReward][SendCoinsFromModuleToAccount] failed. couldn't send coin to cra address.")
	}

	wallet.Reward = sdkmath.LegacyNewDec(0)
	k.SetWallet(ctx, wallet)

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
