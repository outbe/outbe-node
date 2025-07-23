package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/outbe/outbe-node/x/rand/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) CurrentRandomness(goCtx context.Context, req *types.QueryCurrentRandomnessRequest) (*types.QueryCurrentRandomnessResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	state, _ := k.GetPeriod(ctx)

	return &types.QueryCurrentRandomnessResponse{
		Period:     state.CurrentPeriod,
		Randomness: state.CurrentSeed,
	}, nil
}
