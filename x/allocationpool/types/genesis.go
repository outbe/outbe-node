package types

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

func DefaultGenesis() *GenesisState {
	return &GenesisState{
		// this line is used by starport scaffolding # genesis/types/default
		SupplyList:   []Supply{},
		TributeList:  []Tribute{},
		EmissionList: []Emission{},
	}
}
