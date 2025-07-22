package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/outbe/outbe-node/x/rand/types"
)

// InitGenesis initializes the rand module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)
	k.SetPeriod(ctx, *genState.Period)
	for _, commitment := range genState.Commitments {
		k.SetCommitment(ctx, *commitment)
	}
	for _, reveal := range genState.Reveals {
		k.SetReveal(ctx, *reveal)
	}
}

// ExportGenesis returns the rand module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	randState, _ := k.GetPeriod(ctx)
	return &types.GenesisState{
		Params:      k.GetParams(ctx),
		Period:      &randState,
		Commitments: k.GetCommitmentsForPeriod(ctx, randState.CurrentPeriod),
		Reveals:     k.GetRevealsForPeriod(ctx, randState.CurrentPeriod),
	}
}
