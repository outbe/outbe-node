package wasmbinding

import (
	keeper "github.com/outbe/outbe-node/x/rand/keeper"
)

type QueryPlugin struct {
	keeper *keeper.Keeper
}

// NewQueryPlugin returns a reference to a new QueryPlugin.
func NewQueryPlugin(
	rand *keeper.Keeper,
) *QueryPlugin {
	return &QueryPlugin{
		keeper: rand,
	}
}
