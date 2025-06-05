package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/outbe/outbe-node/errors"
	"github.com/outbe/outbe-node/x/gemmint/types"
)

func (k msgServer) AddContractToWhitelist(goCtx context.Context, msg *types.MsgAddContractToWhitelist) (*types.MsgAddContractToWhitelistResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	logger := k.Logger(ctx)

	logger.Info("######### Generating Register Eligible Contract Transaction Started #########")

	// Attempt to retrieve the existing whitelist from storage
	whitelists := k.GetWhitelist(ctx)

	// Create a new eligible contract
	newEligibleContract := &types.EligibleContract{
		ContractAddress: msg.ContractAddress,
		Enabled:         true,
		TargetMint:      0,
		TotalMinted:     0,
		Created:         ctx.BlockTime().String(),
	}

	if len(whitelists) == 0 {
		// If no whitelist exists, create a new one
		newWhitelist := types.Whitelist{
			Creator:           msg.Creator,
			EligibleContracts: []*types.EligibleContract{newEligibleContract},
			TotalMinted:       0,
			Created:           ctx.BlockTime().String(),
		}

		// Save the newly created whitelist to storage
		k.SetWhitelist(ctx, newWhitelist)
		logger.Info("New whitelist created and contract added", "creator", msg.Creator, "contract", msg.ContractAddress)
	} else {
		// Validate that the sender is the whitelist owner
		if whitelists[0].Creator != msg.Creator {
			logger.Error("[AddContractToWhitelist] Unauthorized transaction: sender is not the whitelist owner",
				"sender", msg.Creator,
				"whitelist creator", whitelists[0].Creator,
			)
			return nil, sdkerrors.Wrapf(errortypes.ErrUnauthorized, "sender is not the owner of the whitelist")
		}

		// Check if the contract already exists in the whitelist
		for _, contract := range whitelists[0].EligibleContracts {
			if contract.ContractAddress == msg.ContractAddress {
				logger.Error("[AddContractToWhitelist] Contract already exists in whitelist", "contract", msg.ContractAddress)
				return nil, sdkerrors.Wrapf(errortypes.ErrConflict, "contract already exists in whitelist")
			}
		}

		// Add the new eligible contract to the existing whitelist
		whitelists[0].EligibleContracts = append(whitelists[0].EligibleContracts, newEligibleContract)
		k.SetWhitelist(ctx, whitelists[0])
		logger.Info("Contract successfully added to existing whitelist", "creator", msg.Creator, "contract", msg.ContractAddress)
	}

	logger.Info("######### End of Register Eligible Contract Transaction #########")

	return &types.MsgAddContractToWhitelistResponse{}, nil
}
