package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/outbe/outbe-node/errors"
	"github.com/outbe/outbe-node/x/allocationpool/types"
)

func (k msgServer) RegisterCRA(goCtx context.Context, msg *types.MsgRegisterCRA) (*types.MsgRegisterCRAResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	logger := k.Logger(ctx)

	logger.Info("🔁 Starting submit cra transaction")

	if msg.CraAddress == "" {
		return &types.MsgRegisterCRAResponse{}, sdkerrors.Wrap(errortypes.ErrInvalidAddress, "cra address can not be empty.")
	}

	checkEligibleCU, found := k.GetCRAByCRAAddress(ctx, msg.CraAddress)
	if !found {
		cra := types.CRACU{
			CraAddress: msg.CraAddress,
			Reward:     sdkmath.LegacyNewDec(0),
		}
		k.SetCRA(ctx, cra)
	} else {
		cra := types.CRACU{
			CraAddress: checkEligibleCU.CraAddress,
			Reward:     sdkmath.LegacyNewDec(0),
		}
		k.SetCRA(ctx, cra)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventType,
			sdk.NewAttribute(types.AttributeCraAddress, checkEligibleCU.CraAddress),
		),
	)

	logger.Info("✅ Submitting a cra successfully completed")

	return &types.MsgRegisterCRAResponse{}, nil
}
