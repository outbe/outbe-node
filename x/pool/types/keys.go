package types

import "github.com/cosmos/cosmos-sdk/types/address"

const (
	// ModuleName defines the module name
	ModuleName = "pool"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_pool"
)

var (
	PoolKey        = []byte{0x01}
	TotalSupplyKey = []byte{0x02}
	TributeKey     = []byte("Tribute/")
	MintKey        = []byte{0x04}
	EmissionKey    = []byte{0x12}
	TotalMintedKey = []byte{0x13}
	//    = []byte{0x14}
	//    = []byte{0x15}
	//    = []byte{0x21}
	//    = []byte{0x22}
	//    = []byte{0x23}
	//    = []byte{0x24}
	//    = []byte{0x31}
	//    = []byte{0x32}
)

func GetEmissionKey(id string) []byte {
	return append(EmissionKey, address.MustLengthPrefix([]byte(id))...)
}

func GetPoolKey(id string) []byte {
	return append(PoolKey, address.MustLengthPrefix([]byte(id))...)
}

func GetTributeKey(id string) []byte {
	return append(TributeKey, []byte(id)...)
}

func GetTotalMintedKey(id string) []byte {
	return append(TotalMintedKey, address.MustLengthPrefix([]byte(id))...)
}

func GetTotalSupplyKey(id string) []byte {
	return append(TotalSupplyKey, address.MustLengthPrefix([]byte(id))...)
}

func KeyPrefix(p string) []byte {
	return []byte(p)
}
