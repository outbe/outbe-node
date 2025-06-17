package keeper

import (
	"context"

	"github.com/outbe/outbe-node/x/pool/types"
)

func (k Keeper) SetTotalMinted(ctx context.Context, minted types.Minted) error {
	store := k.storeService.OpenKVStore(ctx)
	b := k.cdc.MustMarshal(&minted)
	return store.Set(types.GetTotalMintedKey("total_minted"), b)
}

func (k Keeper) GetMintedAmount(ctx context.Context) (val types.Minted, found bool) {
	store := k.storeService.OpenKVStore(ctx)
	totalMintedKey := types.GetTotalMintedKey("total_minted")
	b, err := store.Get(totalMintedKey)

	if b == nil || err != nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}
