package types

import (
	fmt "fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	proto "github.com/cosmos/gogoproto/proto"
)

var _ paramtypes.ParamSet = (*Params)(nil)

const (
	DefaultMinAPR = 4 // 4% APR
)

func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// type Params struct {
// 	MinAPR uint32 // expressed in percentage (e.g., 4 for 4%)
// }

// ParamSetPairs implements the paramtypes.ParamSet interface.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair([]byte("MinAPR"), &p.Decay, validateEra0Decay),
		paramtypes.NewParamSetPair([]byte("MinAPR"), &p.InitialRate, validateEra0InitialRate),
	}
}

// validateMinAPR validates the MinAPR parameter.
func validateEra0Decay(i interface{}) error {
	_, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateEra0InitialRate(i interface{}) error {
	_, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func (m *Params) String() string { return proto.CompactTextString(m) }

func DefaultParams() Params {
	return Params{
		Decay:       "0.00000006",
		InitialRate: "65536",
	}
}
