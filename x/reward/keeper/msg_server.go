package keeper

// var _ types.MsgServer = msgServer{}

// msgServer is a wrapper of Keeper.
type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the x/mint MsgServer interface.
// func NewMsgServerImpl(k Keeper) types.MsgServer {
// 	return &msgServer{
// 		Keeper: k,
// 	}
// }
