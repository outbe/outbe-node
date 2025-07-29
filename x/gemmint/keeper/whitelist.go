package keeper

import (
	"context"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/outbe/outbe-node/x/gemmint/types"
)

func (k Keeper) SetWhitelist(ctx context.Context, whitelist types.Whitelist) error {
	store := k.storeService.OpenKVStore(ctx)
	b := k.cdc.MustMarshal(&whitelist)
	return store.Set(types.GetWhitelistKey(whitelist.ContractAddress), b)
}

func (k Keeper) GetContractByAddress(ctx context.Context, address string) (contract types.Whitelist, found bool) {
	store := k.storeService.OpenKVStore(ctx)
	contractKey := types.GetWhitelistKey(address)

	b, err := store.Get(contractKey)

	if b == nil || err != nil {
		return contract, false
	}

	k.cdc.MustUnmarshal(b, &contract)
	return contract, true
}

func (k Keeper) GetWhitelist(ctx context.Context) (list []types.Whitelist) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iterator := storetypes.KVStorePrefixIterator(store, types.WhitelistKey)

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Whitelist
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}
	return
}

func (k Keeper) IsEligibleSmartContract(ctx context.Context, contractAddress string) bool {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iterator := storetypes.KVStorePrefixIterator(store, types.WhitelistKey)

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var whitelist types.Whitelist
		if err := k.cdc.Unmarshal(iterator.Value(), &whitelist); err != nil {
			return false
		}

		if whitelist.ContractAddress == contractAddress && whitelist.Enabled {
			return true
		}
	}

	return false
}
