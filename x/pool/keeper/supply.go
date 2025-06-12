package keeper

import (
	"context"
	"fmt"
	"strconv"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/outbe/outbe-node/x/pool/types"
)

func (k Keeper) SetSupply(ctx context.Context, supply types.Supply) error {
	if _, err := strconv.ParseUint(supply.TotalSupply, 10, 64); err != nil {
		return fmt.Errorf("invalid total supply: %w", err)
	}

	store := k.storeService.OpenKVStore(ctx)
	b := k.cdc.MustMarshal(&supply)
	if err := store.Set(types.TotalSupplyKey, b); err != nil {
		return fmt.Errorf("failed to store supply: %w", err)
	}
	return nil
}

func (k Keeper) TotalSupplyAll(ctx context.Context) (list []types.Supply) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iterator := storetypes.KVStorePrefixIterator(store, types.TotalSupplyKey)

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Supply
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}
	return
}
