package keeper

import (
	"context"
	"fmt"

	gemminttypes "github.com/outbe/outbe-node/x/gemmint/types"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	errortypes "github.com/outbe/outbe-node/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdkerrors "cosmossdk.io/errors"
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

// Inflation returns minter.Inflation of the mint module.
func (q queryServer) Inflation(ctx context.Context, _ *gemminttypes.QueryInflationRequest) (*gemminttypes.QueryInflationResponse, error) {
	minter, err := q.k.Minter.Get(ctx)
	if err != nil {
		return nil, err
	}

	return &gemminttypes.QueryInflationResponse{Inflation: minter.Inflation}, nil
}

// AnnualProvisions returns minter.AnnualProvisions of the mint module.
func (q queryServer) AnnualProvisions(ctx context.Context, _ *gemminttypes.QueryAnnualProvisionsRequest) (*gemminttypes.QueryAnnualProvisionsResponse, error) {
	minter, err := q.k.Minter.Get(ctx)
	if err != nil {
		return nil, err
	}

	return &gemminttypes.QueryAnnualProvisionsResponse{AnnualProvisions: minter.AnnualProvisions}, nil
}

func (k queryServer) Whitelist(c context.Context, req *gemminttypes.QueryWhitelistRequest) (*gemminttypes.QueryWhitelistResponse, error) {

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "[Whitelist] failed. Invalid request.")
	}

	var whitelist []gemminttypes.Whitelist
	ctx := sdk.UnwrapSDKContext(c)

	store := runtime.KVStoreAdapter(k.k.storeService.OpenKVStore(ctx))
	whitelistStore := prefix.NewStore(store, gemminttypes.WhitelistKey)

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
		var whitelist_details gemminttypes.Whitelist
		if err := k.k.cdc.Unmarshal(value, &whitelist_details); err != nil {
			return sdkerrors.Wrap(errortypes.ErrJSONUnmarshal, "[Whitelist][Unmarshal] failed. Couldn't parse the Whitelist data encoded.")
		}
		whitelist = append(whitelist, whitelist_details)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, "[Whitelist] failed. Couldn't find a Whitelist infos.")
	}

	return &gemminttypes.QueryWhitelistResponse{Whitelist: whitelist, Pagination: pageRes}, nil
}

func (q queryServer) Minted(c context.Context, req *gemminttypes.QueryMintedRequest) (*gemminttypes.QueryMintedResponse, error) {

	fmt.Println("query Minted is started------------->")

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "[Minted] failed. Invalid request.")
	}

	minted, found := q.k.GetTotalMinted(c)

	fmt.Println("111111111111111111111--minted", minted)

	if !found {
		minted.TotalMinted = sdkmath.LegacyNewDecFromInt(sdkmath.NewInt(0))
	}

	return &gemminttypes.QueryMintedResponse{TotalMinted: minted}, nil
}
