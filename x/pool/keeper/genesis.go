package keeper

import (
	"github.com/outbe/outbe-node/x/pool/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {

	for _, elem := range genState.SupplyList {
		k.SetSupply(ctx, elem)
	}
	for _, elem := range genState.TributeList {
		k.SetTribute(ctx, elem)
	}
	for _, elem := range genState.EmissionList {
		k.SetEmission(ctx, elem)
	}
}
