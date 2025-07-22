package keeper

import (
	"context"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/outbe/outbe-node/x/rand/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) Commitment(c context.Context, req *types.QueryCommitmentRequest) (*types.QueryCommitmentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	i, _ := strconv.ParseUint(req.Period, 10, 64)

	commitment, _ := k.GetCommitment(ctx, i, req.Validator)

	return &types.QueryCommitmentResponse{Commitment: &commitment}, nil
}

func (k Keeper) Commitments(c context.Context, req *types.QueryCommitmentsRequest) (*types.QueryCommitmentsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	return &types.QueryCommitmentsResponse{Commitments: k.GetCommitments(ctx)}, nil
}
