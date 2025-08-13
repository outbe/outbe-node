package keeper

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/outbe/outbe-node/x/allocationpool/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) GetDailyEmission(goCtx context.Context, req *types.QueryDailyEmissionRequest) (*types.QueryDailyEmissionResponse, error) {

	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "[GetDailyEmission:] invalid request - request cannot be nil")
	}

	dailyEmission, found := k.GetDailyEmissionAmount(ctx)
	if !found {
		return nil, errors.New("[GetDailyEmission][GetDailyEmissionAmount] failed to fetch daily emission")
	}

	return &types.QueryDailyEmissionResponse{CraDailyEmission: &dailyEmission}, nil
}
