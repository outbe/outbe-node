package keeper

import (
	"context"

	"cosmossdk.io/core/address"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx context.Context, senderPool, recipientPool string, amt sdk.Coins) error
	UndelegateCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	DelegateCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error

	BurnCoins(ctx context.Context, name string, amt sdk.Coins) error
}

type StakingKeeper interface {
	ValidatorAddressCodec() address.Codec
	ConsensusAddressCodec() address.Codec
	// iterate through validators by operator address, execute func for each validator
	IterateValidators(context.Context,
		func(index int64, validator stakingtypes.ValidatorI) (stop bool)) error

	Validator(context.Context, sdk.ValAddress) (stakingtypes.ValidatorI, error)            // get a particular validator by operator address
	ValidatorByConsAddr(context.Context, sdk.ConsAddress) (stakingtypes.ValidatorI, error) // get a particular validator by consensus address

	// Delegation allows for getting a particular delegation for a given validator
	// and delegator outside the scope of the staking module.
	Delegation(context.Context, sdk.AccAddress, sdk.ValAddress) (stakingtypes.DelegationI, error)

	IterateDelegations(ctx context.Context, delegator sdk.AccAddress,
		fn func(index int64, delegation stakingtypes.DelegationI) (stop bool)) error

	GetAllSDKDelegations(ctx context.Context) ([]stakingtypes.Delegation, error)
	GetAllValidators(ctx context.Context) ([]stakingtypes.Validator, error)
	GetAllDelegatorDelegations(ctx context.Context, delegator sdk.AccAddress) ([]stakingtypes.Delegation, error)
}

type RewardKeeper interface {
	GetParams(ctx sdk.Context) (params rewardtypes.Params)
	CalculateValidatorFeeShare(transactionFees, selfBondedTokens, totalSelfBondedTokens sdkmath.LegacyDec) (sdkmath.LegacyDec, error)
	CalculateMinimumAPRReward(selfBondedTokens, apr sdkmath.LegacyDec, blocksPerYear int64) (sdkmath.LegacyDec, error)
}

type WrappedBaseKeeper struct {
	distributionkeeper.Keeper

	ak AccountKeeper
	bk BankKeeper
	sk StakingKeeper
	rk RewardKeeper
}

func NewWrappedBaseKeeper(
	sk distributionkeeper.Keeper,

	ak AccountKeeper,
	bk BankKeeper,
	stakingKeeper StakingKeeper,
	rewardKeeper RewardKeeper,
) WrappedBaseKeeper {
	return WrappedBaseKeeper{
		Keeper: sk,

		ak: ak,
		bk: bk,
		sk: stakingKeeper,
		rk: rewardKeeper,
	}
}

func (k WrappedBaseKeeper) UnwrapKeeper() distributionkeeper.Keeper {
	return k.Keeper
}
