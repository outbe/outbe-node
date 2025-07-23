package keeper

import (
	"crypto/sha256"
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) GenerateRandomValueBinary(ctx sdk.Context, valAddr sdk.ValAddress, period uint64) []byte {
	input := append(valAddr.Bytes(), make([]byte, 16)...)
	binary.BigEndian.PutUint64(input[len(valAddr):], uint64(ctx.BlockHeight()))
	binary.BigEndian.PutUint64(input[len(valAddr)+8:], period)
	hash := sha256.Sum256(input)
	return hash[:]
}

// func ComputeHash(randomValue []byte) []byte {
// 	hash := sha256.Sum256([]byte(randomValue))
// 	return hash[:]
// }

func ComputeHash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}
