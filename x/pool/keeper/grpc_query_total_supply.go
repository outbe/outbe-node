package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/outbe/outbe-node/errors"
	"github.com/outbe/outbe-node/x/pool/types"
)

func (k Keeper) GetTotalSupply(goCtx context.Context, req *types.QueryTotalSupplyRequest) (*types.QueryTotalSupplyResponse, error) {
	if req == nil {
		return nil, errortypes.ErrInvalidRequest
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	totalSupply := k.TotalSupplyAll(ctx)
	if len(totalSupply) == 0 {
		supply := types.Supply{
			TotalSupply: "0",
		}
		return &types.QueryTotalSupplyResponse{TotalSupply: &supply}, nil
	}
	return &types.QueryTotalSupplyResponse{TotalSupply: &types.Supply{TotalSupply: totalSupply[0].TotalSupply}}, nil
}
