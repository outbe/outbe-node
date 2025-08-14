package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	errortypes "github.com/outbe/outbe-node/errors"
	"github.com/outbe/outbe-node/x/cra/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = queryServer{}

func NewQueryServerImpl(k Keeper) types.QueryServer {
	return queryServer{k}
}

type queryServer struct {
	k Keeper
}

// Params returns params of the mint module.
func (q queryServer) Params(ctx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	params, err := q.k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryParamsResponse{Params: params}, nil
}

func (q queryServer) AllCRAs(c context.Context, req *types.QueryAllCRAsRequest) (*types.QueryAllCRAsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "[AllCRAs] failed. Invalid request.")
	}

	var cras []types.CRACU
	ctx := sdk.UnwrapSDKContext(c)

	store := runtime.KVStoreAdapter(q.k.storeService.OpenKVStore(ctx))
	cuStore := prefix.NewStore(store, types.CraKey)

	pagination := req.Pagination
	if pagination == nil {
		pagination = &query.PageRequest{Limit: 100, CountTotal: true}
	} else {
		if pagination.Limit == 0 || pagination.Limit > 1000 {
			pagination.Limit = 1000
		}
		pagination.CountTotal = true
	}

	pageRes, err := query.Paginate(cuStore, pagination, func(key []byte, value []byte) error {
		var cra types.CRACU
		if err := q.k.cdc.Unmarshal(value, &cra); err != nil {
			return sdkerrors.Wrap(errortypes.ErrJSONUnmarshal, "[AllCRAs][Unmarshal] failed. Couldn't parse the cu data encoded.")
		}
		cras = append(cras, cra)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "[AllCRAs] failed. Couldn't find an valid cu.")
	}
	return &types.QueryAllCRAsResponse{Cras: cras, Pagination: pageRes}, nil
}

func (q queryServer) AllWallets(c context.Context, req *types.QueryAllWalletsRequest) (*types.QueryAllWalletsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "[AllWallets] failed. Invalid request.")
	}

	var wallets []types.Wallet
	ctx := sdk.UnwrapSDKContext(c)

	store := runtime.KVStoreAdapter(q.k.storeService.OpenKVStore(ctx))
	cuStore := prefix.NewStore(store, types.WalletKey)

	pagination := req.Pagination
	if pagination == nil {
		pagination = &query.PageRequest{Limit: 100, CountTotal: true}
	} else {
		if pagination.Limit == 0 || pagination.Limit > 1000 {
			pagination.Limit = 1000
		}
		pagination.CountTotal = true
	}

	pageRes, err := query.Paginate(cuStore, pagination, func(key []byte, value []byte) error {
		var wallet types.Wallet
		if err := q.k.cdc.Unmarshal(value, &wallet); err != nil {
			return sdkerrors.Wrap(errortypes.ErrJSONUnmarshal, "[AllWallets][Unmarshal] failed. Couldn't parse the cu data encoded.")
		}
		wallets = append(wallets, wallet)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "[AllWallets] failed. Couldn't find an valid cu.")
	}
	return &types.QueryAllWalletsResponse{Wallets: wallets, Pagination: pageRes}, nil
}
