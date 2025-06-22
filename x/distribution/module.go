package distribution

import (
	"context"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/bank/exported"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingKeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	gemdistributionkeeper "github.com/outbe/outbe-node/x/distribution/keeper"
	gemdistributiontypes "github.com/outbe/outbe-node/x/distribution/types"
	rewardKeeper "github.com/outbe/outbe-node/x/reward/keeper"
)

var (
	_ appmodule.HasBeginBlocker = AppModule{}
)

type AppModule struct {
	distribution.AppModule

	keeper   distributionkeeper.Keeper
	wKeeper  gemdistributionkeeper.WrappedBaseKeeper
	subspace exported.Subspace
}

// NewAppModule creates a new AppModule object
func NewAppModule(module distribution.AppModule, keeper distributionkeeper.Keeper, accountKeeper gemdistributiontypes.AccountKeeper, bankKeeper gemdistributiontypes.BankKeeper, ss exported.Subspace, distributionkeeper distributionkeeper.Keeper, stakingKeeper stakingKeeper.Keeper, rewardKeeper rewardKeeper.Keeper) AppModule {
	wrappedBankKeeper := gemdistributionkeeper.NewWrappedBaseKeeper(keeper, accountKeeper, bankKeeper, stakingKeeper, rewardKeeper)

	return AppModule{
		AppModule: module,
		keeper:    keeper,
		wKeeper:   wrappedBankKeeper,
		subspace:  ss,
	}
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	distributiontypes.RegisterMsgServer(cfg.MsgServer(), gemdistributionkeeper.NewMsgServerImpl(am.wKeeper))
	querier := distributionkeeper.Querier{Keeper: am.keeper}
	distributiontypes.RegisterQueryServer(cfg.QueryServer(), querier)
}

// BeginBlock returns the begin blocker for the staking module.
func (am AppModule) BeginBlock(ctx context.Context) error {
	return am.wKeeper.BeginBlocker(ctx)
}
