package types

import (
	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/types/address"
)

var (
	ParamsKey = collections.NewPrefix(8)
	CraKey    = []byte{0x01}
	WalletKey = []byte{0x02}
)

const (
	// module name
	ModuleName = "cra"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_cra"
)

func GetCRAKey(id string) []byte {
	return append(CraKey, address.MustLengthPrefix([]byte(id))...)
}

func GetWalletKey(id string) []byte {
	return append(WalletKey, address.MustLengthPrefix([]byte(id))...)
}
