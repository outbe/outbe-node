package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"

	//sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/outbe/outbe-node/x/cra/types"
)

func (k Keeper) SetCRA(ctx context.Context, cra types.CRA) error {
	store := k.storeService.OpenKVStore(ctx)
	b := k.cdc.MustMarshal(&cra)
	key := types.GetCRAKey(cra.CraAddress)
	return store.Set(key, b)
}

func (k Keeper) SetWallet(ctx context.Context, cra types.Wallet) error {
	store := k.storeService.OpenKVStore(ctx)
	b := k.cdc.MustMarshal(&cra)
	key := types.GetWalletKey(cra.Address)
	return store.Set(key, b)
}

func (k Keeper) GetCRAByCRAAddress(ctx context.Context, address string) (cra types.CRA, found bool) {
	store := k.storeService.OpenKVStore(ctx)
	craKey := types.GetCRAKey(address)
	b, err := store.Get(craKey)

	if b == nil || err != nil {
		return cra, false
	}

	k.cdc.MustUnmarshal(b, &cra)
	return cra, true
}

func (k Keeper) GetWalletByWalletAddress(ctx context.Context, address string) (wallte types.Wallet, found bool) {
	store := k.storeService.OpenKVStore(ctx)
	walletKey := types.GetWalletKey(address)
	b, err := store.Get(walletKey)

	if b == nil || err != nil {
		return wallte, false
	}

	k.cdc.MustUnmarshal(b, &wallte)
	return wallte, true
}

func (k Keeper) GetCRAAll(ctx context.Context) (list []types.CRA) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	tributeStore := prefix.NewStore(store, types.CraKey)
	iterator := tributeStore.Iterator(nil, nil)

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.CRA
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}
	return
}

func (k Keeper) GetWalletAll(ctx context.Context) (list []types.Wallet) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	tributeStore := prefix.NewStore(store, types.WalletKey)
	iterator := tributeStore.Iterator(nil, nil)

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Wallet
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}
	return
}
