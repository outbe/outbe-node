package keeper

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/outbe/outbe-node/x/pool/types"
)

func (k Keeper) GetTotalMinted(goCtx context.Context, req *types.QueryTotalMintedRequest) (*types.QueryTotalMintedResponse, error) {

	ctx := sdk.UnwrapSDKContext(goCtx)

	totalMinted, found := k.GetMintedAmount(ctx)

	if !found {
		return nil, errors.New("[GetTotalMinted][GetMintedAmount] failed. Fetching total minted amount failed")
	}

	return &types.QueryTotalMintedResponse{TotalMinted: &totalMinted}, nil
}
