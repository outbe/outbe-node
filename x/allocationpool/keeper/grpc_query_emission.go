package keeper

import (
	"context"
	"errors"
	"strconv"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	errortypes "github.com/outbe/outbe-node/errors"
	"github.com/outbe/outbe-node/x/allocationpool/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) GetEmission(goCtx context.Context, req *types.QueryEmissionRequest) (*types.QueryEmissionResponse, error) {

	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "[GetEmission:] invalid request - request cannot be nil")
	}

	totalEmission, found := k.GetEmissionState(ctx)
	if !found {
		return nil, errors.New("[GetEmission][GetTotalEmission] failed to fetch total emission")
	}

	return &types.QueryEmissionResponse{Emission: &totalEmission}, nil
}

func (k Keeper) EmissionAll(c context.Context, req *types.QueryAllEmissionRequest) (*types.QueryAllEmissionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "[EmissionAll] failed. Invalid request.")
	}

	var emissions []types.Emission
	ctx := sdk.UnwrapSDKContext(c)

	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	emissionStore := prefix.NewStore(store, types.EmissionKey)

	pagination := req.Pagination
	if pagination == nil {
		pagination = &query.PageRequest{Limit: 100, CountTotal: true}
	} else {
		if pagination.Limit == 0 || pagination.Limit > 1000 {
			pagination.Limit = 1000
		}
		pagination.CountTotal = true
	}

	pageRes, err := query.Paginate(emissionStore, pagination, func(key []byte, value []byte) error {
		var emissionState types.Emission
		if err := k.cdc.Unmarshal(value, &emissionState); err != nil {
			return sdkerrors.Wrap(errortypes.ErrJSONUnmarshal, "[EmissionAll][Unmarshal] failed. Couldn't parse the Motus data encoded.")
		}
		emissions = append(emissions, emissionState)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "[EmissionAll] failed. Couldn't find an valid emission.")
	}

	return &types.QueryAllEmissionResponse{Emissions: emissions, Pagination: pageRes}, nil
}

func (k Keeper) GetEmissionEntity(goCtx context.Context, req *types.QueryEmissionEntityRequest) (*types.QueryEmissionEntityResponse, error) {

	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "[GetEmission:] invalid request - request cannot be nil")
	}

	totalEmission, found := k.GetEmissionEntityPerBlock(ctx, strconv.FormatInt(req.BlockNumber, 10))
	if !found {
		return nil, errors.New("[GetEmissionEntity][GetEmissionEntityPerBlock] failed to fetch emission per block")
	}

	return &types.QueryEmissionEntityResponse{Emission: &totalEmission}, nil
}
