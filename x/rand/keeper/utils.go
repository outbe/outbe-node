package keeper

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) GenerateRandomValue(validatorAddress types.ValAddress, period uint64) string {
	seedData := append(validatorAddress.Bytes(), binary.BigEndian.AppendUint64(nil, period)...)
	hash := sha256.Sum256(seedData)
	return fmt.Sprintf("%x", hash[:])
}

// ComputeHash creates a SHA-256 hash of the random value
func (k Keeper) ComputeHash(randomValue string) []byte {
	hash := sha256.Sum256([]byte(randomValue))
	return hash[:]
}

// CompareRandomWithHash verifies if the random value matches the stored hash
func (k Keeper) CompareRandomWithHash(randomValue string, storedHash []byte) bool {
	computedHash := k.ComputeHash(randomValue)
	return bytes.Equal(storedHash, computedHash)
}

func (k Keeper) ConvertRandomValueToBytes(randomValue string) ([]byte, error) {
	return hex.DecodeString(randomValue)
}
