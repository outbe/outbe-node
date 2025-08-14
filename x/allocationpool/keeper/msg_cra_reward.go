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

func (k msgServer) CRAReward(goCtx context.Context, msg *types.MsgCRAReward) (*types.MsgCRARewardResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	logger := k.Logger(ctx)

	logger.Info("🔁 Starting cra reward transaction")

	if msg.CraAddress == "" {
		return &types.MsgCRARewardResponse{}, sdkerrors.Wrap(errortypes.ErrInvalidAddress, "cra address can not be empty.")
	}

	cra, found := k.GetCRAByCRAAddress(ctx, msg.CraAddress)
	if !found {
		return &types.MsgCRARewardResponse{}, sdkerrors.Wrap(errortypes.ErrInvalidAddress, "[CRAReward][GetCRAByCRAAddress] failed. couldn't fetch a valid cra.")
	}

	err := k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		sdk.AccAddress(cra.CraAddress),
		sdk.NewCoins(sdk.NewCoin(params.BondDenom, sdkmath.Int(cra.Reward))),
	)
	if err != nil {
		return &types.MsgCRARewardResponse{}, sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[CRAReward][SendCoinsFromModuleToAccount] failed. couldn't send coin to cra address.")
	}

	cra.Reward = sdkmath.LegacyNewDec(0)
	k.SetCRA(ctx, cra)

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventType,
			sdk.NewAttribute(types.AttributeCraAddress, cra.CraAddress),
		),
	)

	logger.Info("✅ Submitting a cra reward transaction successfully completed")

	return &types.MsgCRARewardResponse{}, nil
}
