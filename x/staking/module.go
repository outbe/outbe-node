package staking

import (
	"context"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/bank/exported"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	gemstakingkeeper "github.com/outbe/outbe-node/x/staking/keeper"
	gemstakingtypes "github.com/outbe/outbe-node/x/staking/types"
)

var (
	_ appmodule.HasBeginBlocker = AppModule{}
)

type AppModule struct {
	staking.AppModule

	keeper   stakingkeeper.Keeper
	wKeeper  gemstakingkeeper.WrappedBaseKeeper
	subspace exported.Subspace
}

// NewAppModule creates a new AppModule object
func NewAppModule(module staking.AppModule, keeper stakingkeeper.Keeper, accountKeeper gemstakingtypes.AccountKeeper, bankKeeper gemstakingtypes.BankKeeper, ss exported.Subspace, stakingkeeper stakingkeeper.Keeper, rewardKeeper gemstakingtypes.RewardKeeper) AppModule {
	wrappedBankKeeper := gemstakingkeeper.NewWrappedBaseKeeper(keeper, accountKeeper, bankKeeper, rewardKeeper)
	return AppModule{
		AppModule: module,
		keeper:    keeper,
		wKeeper:   wrappedBankKeeper,
		subspace:  ss,
	}
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	stakingtypes.RegisterMsgServer(cfg.MsgServer(), gemstakingkeeper.NewMsgServerImpl(am.wKeeper))
	querier := stakingkeeper.Querier{Keeper: &am.keeper}
	stakingtypes.RegisterQueryServer(cfg.QueryServer(), querier)
}

// BeginBlock returns the begin blocker for the staking module.
func (am AppModule) BeginBlock(ctx context.Context) error {
	return am.wKeeper.BeginBlocker(ctx)
}
