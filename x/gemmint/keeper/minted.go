package keeper

import (
	"context"

	"github.com/outbe/outbe-node/x/gemmint/types"
)

func (k Keeper) SetTotalMinted(ctx context.Context, minted types.Minted) error {
	store := k.storeService.OpenKVStore(ctx)
	b := k.cdc.MustMarshal(&minted)
	return store.Set(types.GetMintedKey("minted"), b)
}

func (k Keeper) GetTotalMinted(ctx context.Context) (val types.Minted, found bool) {
	store := k.storeService.OpenKVStore(ctx)
	mintedKey := types.GetMintedKey("minted")
	b, err := store.Get(mintedKey)

	if b == nil || err != nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}
