package keeper

// import (
// 	"context"

// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	"github.com/outbe/outbe-node/x/cra/types"
// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/status"
// )

// func (k Keeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
// 	if req == nil {
// 		return nil, status.Error(codes.InvalidArgument, "[Params:] invalid request - request cannot be nil")
// 	}

// 	ctx := sdk.UnwrapSDKContext(c)

// 	return &types.QueryParamsResponse{Params: k.GetParams(ctx)}, nil
// }
