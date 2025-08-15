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

func (k msgServer) CRAReward(goCtx context.Context, msg *types.MsgCRAReward) (*types.MsgCRARewardResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	logger := k.Logger(ctx)

	logger.Info("🔁 Starting cra reward transaction")

	if msg.CraAddress == "" {
		return &types.MsgCRARewardResponse{}, sdkerrors.Wrap(errortypes.ErrInvalidAddress, "[CRAReward] cra address can not be empty.")
	}

	if msg.Creator == "" {
		return &types.MsgCRARewardResponse{}, sdkerrors.Wrap(errortypes.ErrInvalidAddress, "[CRAReward] creator can not be empty.")
	}

	cra, found := k.GetCRAByCRAAddress(ctx, msg.CraAddress)
	if !found {
		return &types.MsgCRARewardResponse{}, sdkerrors.Wrap(errortypes.ErrInvalidAddress, "[CRAReward][GetCRAByCRAAddress] failed. couldn't fetch a valid cra.")
	}

	spendableCoin := sdk.NewCoins(sdk.NewCoin(params.BondDenom, cra.Reward.TruncateInt().Sub(sdkmath.NewInt(constants.RewardDefragment))))

	err := k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		sdk.MustAccAddressFromBech32(cra.CraAddress),
		spendableCoin,
	)
	if err != nil {
		return &types.MsgCRARewardResponse{}, sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[CRAReward][SendCoinsFromModuleToAccount] failed. couldn't send coin to cra address.")
	}

	newCra := types.CRA{
		Creator:    cra.Creator,
		CraAddress: cra.CraAddress,
		Reward:     sdkmath.LegacyNewDec(constants.RewardDefragment),
	}
	k.SetCRA(ctx, newCra)

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
