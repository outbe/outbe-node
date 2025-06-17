package keeper

import (
	"github.com/outbe/outbe-node/x/pool/types"
)

type msgServer struct {
	Keeper
}

/** NewMsgServerImpl returns an implementation of the MsgServer interface for the provided Keeper. */
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}
