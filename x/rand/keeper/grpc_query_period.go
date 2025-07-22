package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/outbe/outbe-node/x/rand/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) Period(c context.Context, req *types.QueryPeriodRequest) (*types.QueryPeriodResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	period, _ := k.GetPeriod(ctx)

	return &types.QueryPeriodResponse{Period: &period}, nil
}
