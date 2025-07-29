package keeper

import (
	"context"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/outbe/outbe-node/x/gemmint/types"
)

func (k Keeper) GetAllMinters(ctx context.Context) (list []types.Minter) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iterator := storetypes.KVStorePrefixIterator(store, types.MinterKey)

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Minter
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}
	return
}
