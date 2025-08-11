package keeper

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/outbe/outbe-node/x/allocationpool/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) GetEmission(goCtx context.Context, req *types.QueryEmissionRequest) (*types.QueryEmissionResponse, error) {

	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "[GetEmission:] invalid request - request cannot be nil")
	}

	totalEmission, found := k.GetTotalEmission(ctx)
	if !found {
		return nil, errors.New("[GetEmission][GetTotalEmission] failed to fetch total emission")
	}

	return &types.QueryEmissionResponse{Emission: &totalEmission}, nil
}
