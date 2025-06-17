package pool

import (
	"github.com/outbe/outbe-node/x/pool/keeper"
	"github.com/outbe/outbe-node/x/pool/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	return genesis
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func ValidateGenesis(gs *types.GenesisState) error {
	return nil
}
