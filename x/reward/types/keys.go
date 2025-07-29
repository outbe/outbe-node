package types

import (
	"cosmossdk.io/collections"
)

var (
	// MinterKey is the key to use for the keeper store.
	MinterKey    = collections.NewPrefix(7)
	ParamsKey    = collections.NewPrefix(8)
	WhitelistKey = []byte{0x12}
	MintedKey    = []byte{0x13}
)

const (
	// module name
	ModuleName = "reward"

	// StoreKey is the default store key for mint
	StoreKey = ModuleName
)
