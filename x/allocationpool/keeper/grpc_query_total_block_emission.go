package keeper

import (
	"context"
	"errors"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/outbe/outbe-node/x/allocationpool/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) GetTotalBlockEmission(goCtx context.Context, req *types.QueryTotalBlockEmissionRequest) (*types.QueryTotalBlockEmissionResponse, error) {

	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "[GetBlockEmission:] invalid request - request cannot be nil")
	}

	inflation := k.mintKeeper.GetAllMinters(ctx)[0].Inflation

	if inflation.GT(sdkmath.LegacyNewDecWithPrec(2, 2)) {

		if req.BlockNumber < 0 {
			return nil, errors.New("[GetBlockEmission] invalid block number: cannot be negative")
		}

		tokens, err := k.CalculateExponentialBlockEmission(goCtx, req.BlockNumber)
		if err != nil {
			return nil, errors.New("[GetBlockEmission][CalculateExponentialBlockEmission] failed to calculate exponential block emission")
		}

		return &types.QueryTotalBlockEmissionResponse{BlockEmission: tokens}, nil

	} else {

		tokens, err := k.CalculateFixedBlockEmission(goCtx)
		if err != nil {
			return nil, errors.New("[GetBlockEmission][CalculateFixedBlockEmission] failed to calculate fixed emission")
		}

		return &types.QueryTotalBlockEmissionResponse{BlockEmission: tokens}, nil
	}
}
