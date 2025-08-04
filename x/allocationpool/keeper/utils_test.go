package keeper

import (
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// Simulate the counter storage as a byte slice
func generateTributeID(counterBz []byte) (string, []byte, error) {
	var counter uint64
	if counterBz != nil {
		counter = binary.BigEndian.Uint64(counterBz)
	}

	counter++

	newBz := make([]byte, 8)
	binary.BigEndian.PutUint64(newBz, counter)

	return fmt.Sprintf("tribute-%d", counter), newBz, nil
}

func TestGenerateTributeID(t *testing.T) {
	var counterBz []byte // nil on first run

	// First ID generation
	id, newBz, err := generateTributeID(counterBz)
	require.NoError(t, err)
	require.Equal(t, "tribute-1", id)
	counterBz = newBz

	// Second ID generation
	id, newBz, err = generateTributeID(counterBz)
	require.NoError(t, err)
	require.Equal(t, "tribute-2", id)
	counterBz = newBz

	// Generate more IDs to test counting
	for i := 3; i <= 10; i++ {
		id, newBz, err = generateTributeID(counterBz)
		require.NoError(t, err)
		fmt.Println("id:", id)
		require.Equal(t, fmt.Sprintf("tribute-%d", i), id)
		counterBz = newBz
	}
}
