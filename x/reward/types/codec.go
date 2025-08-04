package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(types.NewInterfaceRegistry())
)

func init() {
	RegisterLegacyAminoCodec(Amino)
	sdk.RegisterLegacyAminoCodec(Amino)
	Amino.Seal()
}

// RegisterLegacyAminoCodec registers concrete types on the LegacyAmino codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(Params{}, "reward/Params", nil)
}

// RegisterInterfaces registers the interfaces types with the interface registry.
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		//&MsgAddContractToWhitelist{},
		//&MsgUpdateParams{},
	)

	//msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
