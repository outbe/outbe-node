package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/outbe/outbe-node/errors"
	"github.com/outbe/outbe-node/x/cra/types"
)

func (k msgServer) RegisterWallet(goCtx context.Context, msg *types.MsgRegisterWallet) (*types.MsgRegisterWalletResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	logger := k.Logger(ctx)

	logger.Info("🔁 Starting registering a valid wallet transaction")

	if msg.Address == "" {
		return &types.MsgRegisterWalletResponse{}, sdkerrors.Wrap(errortypes.ErrInvalidAddress, "[RegisterWallet] wallet address can not be empty.")
	}

	if msg.Creator == "" {
		return &types.MsgRegisterWalletResponse{}, sdkerrors.Wrap(errortypes.ErrInvalidAddress, "[RegisterWallet] creator can not be empty.")
	}

	checkEligibleCU, found := k.GetWalletByWalletAddress(ctx, msg.Address)
	if !found {
		wallet := types.Wallet{
			Creator: msg.Creator,
			Address: msg.Address,
			Reward:  sdkmath.LegacyNewDec(0),
		}
		k.SetWallet(ctx, wallet)
	} else {
		return &types.MsgRegisterWalletResponse{}, sdkerrors.Wrap(errortypes.ErrInvalidAddress, "[RegisterWallet] wallet is already registered.")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventType,
			sdk.NewAttribute(types.AttributeCraAddress, checkEligibleCU.Address),
		),
	)

	logger.Info("✅ Registering a wallet successfully completed")

	return &types.MsgRegisterWalletResponse{}, nil
}
