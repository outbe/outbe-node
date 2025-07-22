package keeper

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/outbe/outbe-node/x/rand/constants"
)

func (k Keeper) GenerateRandomValueString(randValue []byte) string {
	hash := sha256.Sum256(randValue)
	return fmt.Sprintf("%x", hash[:])
}

func (k Keeper) GenerateRandomValueBinary(ctx sdk.Context, valAddr sdk.ValAddress, period uint64) []byte {
	input := append(valAddr.Bytes(), make([]byte, 16)...)
	binary.BigEndian.PutUint64(input[len(valAddr):], uint64(ctx.BlockHeight()))
	binary.BigEndian.PutUint64(input[len(valAddr)+8:], period)
	hash := sha256.Sum256(input)
	return hash[:]
}

// ComputeHash creates a SHA-256 hash of the random value
func (k Keeper) ComputeHash(randomValue []byte) []byte {
	hash := sha256.Sum256([]byte(randomValue))
	return hash[:]
}

// CompareRandomWithHash verifies if the random value matches the stored hash
func (k Keeper) CompareRandomWithHash(randomValue []byte, storedHash []byte) bool {
	computedHash := k.ComputeHash(randomValue)
	return bytes.Equal(storedHash, computedHash)
}

func (k Keeper) ConvertRandomValueToBytes(randomValue string) ([]byte, error) {
	return hex.DecodeString(randomValue)
}

func (k Keeper) LoadByteSliceFromFile(filename string) ([]byte, error) {
	// Read file content
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Clean and parse the content
	content := strings.TrimSpace(string(data))
	content = strings.Trim(content, "[]")
	byteStrs := strings.Fields(content)

	// Convert to []byte
	bytes := make([]byte, 0, len(byteStrs))
	for _, s := range byteStrs {
		n, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("invalid byte value %q: %w", s, err)
		}
		bytes = append(bytes, byte(n))
	}

	return bytes, nil
}

func (k Keeper) WriteByteToTheFile(ctx sdk.Context, r []byte) {
	logger := k.Logger(ctx)
	filePath := constants.FilePath
	data := fmt.Sprintf("%d\n", r)
	if err := os.WriteFile(filePath, []byte(data), 0644); err != nil {
		logger.Error("Failed to write to file", "error", err)
	} else {
		logger.Info("Wrote to file for debugging", "path", filePath)
	}
}
