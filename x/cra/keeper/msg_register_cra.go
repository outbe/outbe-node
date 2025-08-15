package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/outbe/outbe-node/errors"
	"github.com/outbe/outbe-node/x/cra/types"
)

func (k msgServer) RegisterCRA(goCtx context.Context, msg *types.MsgRegisterCRA) (*types.MsgRegisterCRAResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	logger := k.Logger(ctx)

	logger.Info("🔁 Starting registering a valid cra transaction")

	if msg.CraAddress == "" {
		return &types.MsgRegisterCRAResponse{}, sdkerrors.Wrap(errortypes.ErrInvalidAddress, "[RegisterCRA] cra address can not be empty.")
	}

	if msg.Creator == "" {
		return &types.MsgRegisterCRAResponse{}, sdkerrors.Wrap(errortypes.ErrInvalidAddress, "[RegisterCRA] creator can not be empty.")
	}

	checkEligibleCU, found := k.GetCRAByCRAAddress(ctx, msg.CraAddress)

	if !found {
		newCra := types.CRA{
			Creator:    msg.Creator,
			CraAddress: msg.CraAddress,
			Reward:     sdkmath.LegacyNewDec(0),
		}
		k.SetCRA(ctx, newCra)
	} else {
		return &types.MsgRegisterCRAResponse{}, sdkerrors.Wrap(errortypes.ErrInvalidAddress, "[RegisterCRA] cra is already registered.")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventType,
			sdk.NewAttribute(types.AttributeCraAddress, checkEligibleCU.CraAddress),
		),
	)

	logger.Info("✅ Regitering a cra successfully completed")

	return &types.MsgRegisterCRAResponse{}, nil
}
