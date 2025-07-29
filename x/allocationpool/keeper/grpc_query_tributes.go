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
	"github.com/outbe/outbe-node/x/allocationpool/types"
	"google.golang.org/grpc/codes"
)

func (k Keeper) GetTributes(c context.Context, req *types.QueryTributesRequest) (*types.QueryTributesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "[GetTribute:] invalid request - request cannot be nil")
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
			return sdkerrors.Wrap(errortypes.ErrJSONUnmarshal, "[GetTribute:] failed to unmarshal tribute data")
		}
		tributes = append(tributes, tribute)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, "[GetTribute:] failed to find a valid tribute")
	}

	return &types.QueryTributesResponse{Tributes: tributes, Pagination: pageRes}, nil
}
