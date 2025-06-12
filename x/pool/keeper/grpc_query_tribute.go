package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/gogo/status"
	errortypes "github.com/outbe/outbe-node/errors"
	"github.com/outbe/outbe-node/x/pool/types"
	"google.golang.org/grpc/codes"
)

func (k Keeper) GetTribute(c context.Context, req *types.QueryTributeRequest) (*types.QueryTributeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "[GetTribute] failed. Invalid request.")
	}

	var tributes []types.Tribute
	ctx := sdk.UnwrapSDKContext(c)

	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	tributeStore := prefix.NewStore(store, types.TributeKey)

	pagination := req.Pagination
	if pagination == nil {
		pagination = &query.PageRequest{Limit: 100, CountTotal: true}
	} else {
		if pagination.Limit == 0 || pagination.Limit > 1000 {
			pagination.Limit = 1000
		}
		pagination.CountTotal = true
	}

	pageRes, err := query.Paginate(tributeStore, pagination, func(key []byte, value []byte) error {
		var tribute types.Tribute
		if err := k.cdc.Unmarshal(value, &tribute); err != nil {
			return sdkerrors.Wrap(errortypes.ErrJSONUnmarshal, "[GetTribute][Unmarshal] failed. Couldn't parse the tribute data encoded.")
		}
		tributes = append(tributes, tribute)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "[GetTribute] failed. Couldn't find a valid tribute.")
	}

	return &types.QueryTributeResponse{Tribute: tributes, Pagination: pageRes}, nil
}
