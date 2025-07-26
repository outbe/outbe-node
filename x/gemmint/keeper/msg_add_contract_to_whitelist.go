package keeper

import (
	"context"
	"time"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/outbe/outbe-node/errors"
	"github.com/outbe/outbe-node/x/gemmint/types"
)

func (k msgServer) AddContractToWhitelist(goCtx context.Context, msg *types.MsgAddContractToWhitelist) (*types.MsgAddContractToWhitelistResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	logger := k.Logger(ctx)

	logger.Info("🔁 Starting eligible smart contract registeration")

	err := k.ValidateContractAddress(msg.ContractAddress)
	if err != nil {
		return nil, sdkerrors.Wrapf(errortypes.ErrInvalidAddress, "[AddContractToWhitelist] smart contract address is not a valid address")
	}

	isValid := k.IsValidCreator(msg.Creator)
	if !isValid {
		return nil, sdkerrors.Wrapf(errortypes.ErrInvalidType, "[AddContractToWhitelist] creator address is not valid")
	}

	// Attempt to retrieve the existing whitelist from storage
	_, found := k.GetContractByAddress(ctx, msg.ContractAddress)
	if found {
		return nil, sdkerrors.Wrapf(errortypes.ErrInvalidType, "[AddContractToWhitelist] smart contract address is already registerd in whitelist")
	}

	newContract := types.Whitelist{
		Creator:         msg.Creator,
		ContractAddress: msg.ContractAddress,
		Created:         time.Now().UTC().Format(time.RFC3339),
		Enabled:         true,
	}
	k.SetWhitelist(ctx, newContract)

	logger.Info("✅ An eligible smart contract successfully registered",
		"whitelists", newContract,
	)

	return &types.MsgAddContractToWhitelistResponse{}, nil
}
