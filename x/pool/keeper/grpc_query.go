package keeper

import (
	"github.com/outbe/outbe-node/x/pool/types"
)

type QueryServer struct {
	Keeper
}

var _ types.QueryServer = Keeper{}

func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &QueryServer{
		Keeper: keeper,
	}
}
