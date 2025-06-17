package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/outbe/outbe-node/x/pool/types"
)

func (k Keeper) SetTribute(ctx context.Context, tributeDetails types.Tribute) error {
	store := k.storeService.OpenKVStore(ctx)
	b := k.cdc.MustMarshal(&tributeDetails)
	key := types.GetTributeKey(tributeDetails.Id)
	return store.Set(key, b)
}

func (k Keeper) GetTributeAll(ctx context.Context) (list []types.Tribute) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	tributeStore := prefix.NewStore(store, types.TributeKey)
	iterator := tributeStore.Iterator(nil, nil)

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Tribute
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}
	return
}
