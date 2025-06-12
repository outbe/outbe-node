package keeper

import (
	"context"
	"encoding/binary"
	"fmt"
)

// func (k Keeper) GenerateTributeID(ctx context.Context, tribute types.Tribute) (string, error) {
// 	// Combine tribute fields into a unique string
// 	data := fmt.Sprintf("%s:%s:%s:%d",
// 		tribute.Creator,
// 		tribute.ContractAddress,
// 		tribute.RecipientAddress,
// 		tribute.Amount,
// 	)

// 	// Generate a SHA-256 hash of the data
// 	hash := sha256.Sum256([]byte(data))

// 	// Convert to hexadecimal string and take first 20 characters for brevity
// 	id := hex.EncodeToString(hash[:])[:20]

// 	return fmt.Sprintf("tribute-%s", id), nil
// }

func (k Keeper) GenerateTributeID(ctx context.Context) (string, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := []byte("TributeIDCounter")

	// Get current counter
	bz, err := store.Get(key)
	if err != nil {
		return "", err
	}

	var counter uint64
	if bz == nil {
		counter = 0
	} else {
		counter = binary.BigEndian.Uint64(bz)
	}

	// Increment counter
	counter++

	// Save new counter
	newBz := make([]byte, 8)
	binary.BigEndian.PutUint64(newBz, counter)
	if err := store.Set(key, newBz); err != nil {
		return "", err
	}

	// Return counter as string
	return fmt.Sprintf("tribute-%d", counter), nil
}
