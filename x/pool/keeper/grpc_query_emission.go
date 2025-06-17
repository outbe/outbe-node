package keeper

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/outbe/outbe-node/x/pool/types"
)

func (k Keeper) GetEmission(goCtx context.Context, req *types.QueryEmissionRequest) (*types.QueryEmissionResponse, error) {

	ctx := sdk.UnwrapSDKContext(goCtx)

	totalEmission, found := k.GetTotalEmission(ctx)
	if !found {
		return nil, errors.New("[GetEmission][GetTotalEmission] fetching total emission failed")
	}

	return &types.QueryEmissionResponse{Emission: &totalEmission}, nil
}
