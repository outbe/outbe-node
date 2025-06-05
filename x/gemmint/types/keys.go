package types

import (
	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/types/address"
)

var (
	// MinterKey is the key to use for the keeper store.
	MinterKey    = collections.NewPrefix(0)
	ParamsKey    = collections.NewPrefix(1)
	WhitelistKey = []byte{0x12}
	MintedKey    = []byte{0x13}
)

const (
	// module name
	ModuleName = "gemmint"

	// StoreKey is the default store key for mint
	StoreKey = ModuleName
)

func GetWhitelistKey(id string) []byte {
	return append(WhitelistKey, address.MustLengthPrefix([]byte(id))...)
}

func GetMintedKey(id string) []byte {
	return append(MintedKey, address.MustLengthPrefix([]byte(id))...)
}
