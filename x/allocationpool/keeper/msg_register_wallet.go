package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/outbe/outbe-node/errors"
	"github.com/outbe/outbe-node/x/allocationpool/types"
)

func (k msgServer) RegisterWallet(goCtx context.Context, msg *types.MsgRegisterWallet) (*types.MsgRegisterWalletResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	logger := k.Logger(ctx)

	logger.Info("🔁 Starting submit wallet transaction")

	if msg.Address == "" {
		return &types.MsgRegisterWalletResponse{}, sdkerrors.Wrap(errortypes.ErrInvalidAddress, "wallet address can not be empty.")
	}

	checkEligibleCU, found := k.GetWalletByCRAAddress(ctx, msg.Address)
	if !found {
		wallet := types.Wallet{
			Address: msg.Address,
			Reward:  sdkmath.LegacyNewDec(0),
		}
		k.SetWallet(ctx, wallet)
	} else {
		wallet := types.Wallet{
			Address: checkEligibleCU.Address,
			Reward:  sdkmath.LegacyNewDec(0),
		}
		k.SetWallet(ctx, wallet)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventType,
			sdk.NewAttribute(types.AttributeCraAddress, checkEligibleCU.Address),
		),
	)

	logger.Info("✅ Submitting a wallet successfully completed")

	return &types.MsgRegisterWalletResponse{}, nil
}
