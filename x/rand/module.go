package rand

import (
	"context"
	"encoding/json"

	"cosmossdk.io/core/appmodule"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/outbe/outbe-node/x/rand/client/cli"
	"github.com/outbe/outbe-node/x/rand/keeper"
	"github.com/outbe/outbe-node/x/rand/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}

	_ module.HasServices  = AppModule{}
	_ appmodule.AppModule = AppModule{}

	_ module.HasGenesis  = AppModule{}
	_ module.HasServices = AppModule{}
)

// ----------------------------------------------------------------------------
// AppModuleBasic
// ----------------------------------------------------------------------------

// AppModuleBasic implements the AppModuleBasic interface for the capability module.
type AppModuleBasic struct {
	cdc codec.BinaryCodec
}

func NewAppModuleBasic(cdc codec.BinaryCodec) AppModuleBasic {
	return AppModuleBasic{cdc: cdc}
}

func (b AppModuleBasic) RegisterLegacyAminoCodec(amino *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(amino)
}

func (AppModuleBasic) RegisterCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// GetQueryCmd returns no root query command for the wasm module.
func (b AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

func (b AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, serveMux *runtime.ServeMux) {
	err := types.RegisterQueryHandlerClient(context.Background(), serveMux, types.NewQueryClient(clientCtx))
	if err != nil {
		panic(err)
	}
}

var _ appmodule.AppModule = AppModule{}

func (AppModuleBasic) Name() string {
	return types.ModuleName
}

func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

func (b AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

func (b AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

var _ appmodule.AppModule = AppModule{}

type AppModule struct {
	AppModuleBasic
	cdc           codec.Codec
	keeper        keeper.Keeper
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
}

func NewAppModule(
	cdc codec.Codec,

	keeper keeper.Keeper,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(cdc),
		cdc:            cdc,

		keeper:        keeper,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
	}
}

func (am AppModule) IsOnePerModuleType() {
}

func (am AppModule) IsAppModule() {
}

func (AppModule) ConsensusVersion() uint64 { return 3 }

func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(am.keeper))
}

func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// QuerierRoute returns the wasm module's querier route name.
func (AppModule) QuerierRoute() string {
	return types.QuerierRoute
}

func (am AppModule) InitGenesis(ctx sdk.Context, jsonCodec codec.JSONCodec, message json.RawMessage) {
	var genesis types.GenesisState
	jsonCodec.MustUnmarshalJSON(message, &genesis)
	am.keeper.InitGenesis(ctx, genesis)
}

func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(genState)
}

func (am AppModule) ValidateGenesis(marshaler codec.JSONCodec, _ client.TxEncodingConfig, message json.RawMessage) error {
	var genesis types.GenesisState
	err := marshaler.UnmarshalJSON(message, &genesis)
	if err != nil {
		return err
	}
	return nil // am.keeper.ValidateGenesis(&genesis)
}

func (am AppModule) BeginBlock(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return am.keeper.BeginBlocker(sdkCtx)
}

func (am AppModule) EndBlock(context context.Context) error {
	ctx := sdk.UnwrapSDKContext(context)
	return am.keeper.EndBlocker(ctx)
}
