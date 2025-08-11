package keeper

import (
	"context"

	"cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	rewardtypes "github.com/outbe/outbe-node/x/reward/types"
)

type AccountKeeper interface {
	AddressCodec() address.Codec

	IterateAccounts(ctx context.Context, process func(sdk.AccountI) (stop bool))
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI // only used for simulation

	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI

	// TODO remove with genesis 2-phases refactor https://github.com/cosmos/cosmos-sdk/issues/2862
	SetModuleAccount(context.Context, sdk.ModuleAccountI)
}

type BankKeeper interface {
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	LockedCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins

	GetSupply(ctx context.Context, denom string) sdk.Coin

	SendCoinsFromModuleToModule(ctx context.Context, senderPool, recipientPool string, amt sdk.Coins) error
	UndelegateCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	DelegateCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error

	BurnCoins(ctx context.Context, name string, amt sdk.Coins) error
}

type RewardKeeper interface {
	GetParams(ctx sdk.Context) rewardtypes.Params
}

type WrappedBaseKeeper struct {
	stakingkeeper.Keeper

	ak AccountKeeper
	bk BankKeeper
	rk RewardKeeper
}

func NewWrappedBaseKeeper(
	sk stakingkeeper.Keeper,

	ak AccountKeeper,
	bk BankKeeper,
	rk RewardKeeper,
) WrappedBaseKeeper {
	return WrappedBaseKeeper{
		Keeper: sk,

		ak: ak,
		bk: bk,
		rk: rk,
	}
}

func (k WrappedBaseKeeper) UnwrapKeeper() stakingkeeper.Keeper {
	return k.Keeper
}

// func (k WrappedBaseKeeper) UnwrapBaseKeeper() stakingkeeper.Keeper {
// 	return k.Keeper.(stakingkeeper.BaseKeeper)
// }
