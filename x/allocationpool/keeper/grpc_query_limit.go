package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/outbe/outbe-node/x/allocationpool/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) GetLimit(goCtx context.Context, req *types.QueryLimitRequest) (*types.QueryLimitResponse, error) {

	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "[GetEmission:] invalid request - request cannot be nil")
	}

	limit := k.CalculateAnnualEmissionLimit(ctx)

	return &types.QueryLimitResponse{Limit: limit}, nil
}
