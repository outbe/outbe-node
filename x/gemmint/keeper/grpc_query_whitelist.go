package keeper

import (
	"context"
	"fmt"

	"github.com/outbe/outbe-node/x/gemmint/types"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	errortypes "github.com/outbe/outbe-node/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdkerrors "cosmossdk.io/errors"
)

func (k Keeper) Whitelist1(c context.Context, req *types.QueryWhitelistRequest) (*types.QueryWhitelistResponse, error) {

	fmt.Println("query is started------------->")

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "[Whitelist] failed. Invalid request.")
	}

	var whitelists []types.Whitelist
	ctx := sdk.UnwrapSDKContext(c)

	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	whitelistStore := prefix.NewStore(store, types.WhitelistKey)

	pagination := req.Pagination
	if pagination == nil {
		pagination = &query.PageRequest{Limit: 100, CountTotal: true}
	} else {
		if pagination.Limit == 0 || pagination.Limit > 1000 {
			pagination.Limit = 1000
		}
		pagination.CountTotal = true
	}

	pageRes, err := query.Paginate(whitelistStore, pagination, func(key []byte, value []byte) error {
		var whitelist_details types.Whitelist
		if err := k.cdc.Unmarshal(value, &whitelist_details); err != nil {
			return sdkerrors.Wrap(errortypes.ErrJSONUnmarshal, "[Whitelist][Unmarshal] failed. Couldn't parse the Whitelist data encoded.")
		}
		fmt.Println("111111111111111111111--whitelist_details", whitelist_details)
		whitelists = append(whitelists, whitelist_details)
		return nil
	})

	fmt.Println("111111111111111111111--whitelist", whitelists)

	if err != nil {
		return nil, status.Error(codes.Internal, "[Whitelist] failed. Couldn't find a Whitelist infos.")
	}

	return &types.QueryWhitelistResponse{Whitelist: whitelists, Pagination: pageRes}, nil
}
