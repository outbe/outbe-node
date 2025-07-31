package keeper

import (
	"context"

	gemminttypes "github.com/outbe/outbe-node/x/reward/types"
)

var _ gemminttypes.QueryServer = queryServer{}

func NewQueryServerImpl(k Keeper) gemminttypes.QueryServer {
	return queryServer{k}
}

type queryServer struct {
	k Keeper
}

// Params returns params of the mint module.
func (q queryServer) Params(ctx context.Context, _ *gemminttypes.QueryParamsRequest) (*gemminttypes.QueryParamsResponse, error) {
	params, err := q.k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &gemminttypes.QueryParamsResponse{Params: params}, nil
}
