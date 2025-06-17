package wasmbinding

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	poolkeeper "github.com/outbe/outbe-node/x/pool/keeper"
)

func RegisterCustomPlugins(
	bank *bankkeeper.BaseKeeper,
	pool *poolkeeper.Keeper,
) []wasmkeeper.Option {

	wasmQueryPlugin := NewQueryPlugin(pool)

	queryPluginOpt := wasmkeeper.WithQueryPlugins(&wasmkeeper.QueryPlugins{
		Custom: CustomQuerier(wasmQueryPlugin),
	})

	messengerDecoratorOpt := wasmkeeper.WithMessageHandlerDecorator(
		CustomMessageDecorator(bank, pool),
	)

	return []wasmkeeper.Option{
		queryPluginOpt,
		messengerDecoratorOpt,
	}
}
