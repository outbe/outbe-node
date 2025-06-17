package keeper

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/outbe/outbe-node/x/pool/constants"
	"github.com/outbe/outbe-node/x/pool/types"
)

func (k Keeper) GetBlockEmission(goCtx context.Context, req *types.QueryBlockEmissionRequest) (*types.QueryBlockEmissionResponse, error) {

	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, errors.New("[GetBlockEmission] failed. request is nil")
	}

	if ctx.BlockHeight() < constants.TransitionBlockNumber {

		if req.BlockNumber < 0 {
			return nil, errors.New("blocknumber is 0")
		}

		tokens, err := k.CalculateExponentialBlockEmission(req.BlockNumber)
		if err != nil {
			return nil, errors.New("[GetBlockEmission][CalculateExponentialBlockEmission] failed.CalculateExponentialTokens failed")
		}
		return &types.QueryBlockEmissionResponse{BlockEmission: tokens}, nil
	}

	tokens, err := k.CalculateFixedBlockEmission(goCtx)
	if err != nil {
		return nil, errors.New("[GetBlockEmission][CalculateFixedBlockEmission] failed. CalculateFixedTokens failed")
	}

	return &types.QueryBlockEmissionResponse{BlockEmission: tokens}, nil
}
