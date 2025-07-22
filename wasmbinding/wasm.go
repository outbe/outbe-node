package wasmbinding

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	randkeeper "github.com/outbe/outbe-node/x/rand/keeper"
)

func RegisterCustomPlugins(
	bank *bankkeeper.BaseKeeper,
	gemmint *randkeeper.Keeper,
) []wasmkeeper.Option {

	wasmQueryPlugin := NewQueryPlugin(gemmint)

	queryPluginOpt := wasmkeeper.WithQueryPlugins(&wasmkeeper.QueryPlugins{
		Custom: CustomQuerier(wasmQueryPlugin),
	})

	// messengerDecoratorOpt := wasmkeeper.WithMessageHandlerDecorator(
	// 	CustomMessageDecorator(bank, gemmint),
	// )

	return []wasmkeeper.Option{
		queryPluginOpt,
		//messengerDecoratorOpt,
	}
}
