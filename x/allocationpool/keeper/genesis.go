package keeper

import (
	"github.com/outbe/outbe-node/x/allocationpool/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {

	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}

	for _, elem := range genState.TributeList {
		k.SetTribute(ctx, elem)
	}

	for _, elem := range genState.EmissionList {
		k.SetEmission(ctx, elem)
	}
}
