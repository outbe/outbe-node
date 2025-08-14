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
	cdc.RegisterConcrete(&MsgRegisterCRA{}, "cra/MsgRegisterCRA", nil)
	cdc.RegisterConcrete(&MsgRegisterWallet{}, "cra/MsgRegisterWallet", nil)
	cdc.RegisterConcrete(&MsgCRAReward{}, "cra/MsgCRAReward", nil)
	cdc.RegisterConcrete(&MsgWalletReward{}, "cra/MsgWalletReward", nil)
}

// RegisterInterfaces registers the interfaces types with the interface registry.
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRegisterCRA{},
		&MsgRegisterWallet{},
		&MsgCRAReward{},
		&MsgWalletReward{},
	)

	//msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
