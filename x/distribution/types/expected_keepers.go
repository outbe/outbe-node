package types

import (
	"context"

	"cosmossdk.io/core/address"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	allocationpooltypes "github.com/outbe/outbe-node/x/allocationpool/types"
	cratypes "github.com/outbe/outbe-node/x/cra/types"
	gemminttypes "github.com/outbe/outbe-node/x/gemmint/types"
	rewardtypes "github.com/outbe/outbe-node/x/reward/types"
)

// AccountKeeper defines the account contract that must be fulfilled when
// creating a x/bank keeper.
type AccountKeeper interface {
	AddressCodec() address.Codec
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx context.Context, name string) sdk.ModuleAccountI
	// TODO remove with genesis 2-phases refactor https://github.com/cosmos/cosmos-sdk/issues/2862
	SetModuleAccount(context.Context, sdk.ModuleAccountI)
	//IterateAccounts(ctx context.Context, process func(sdk.AccountI) bool)
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

type BankKeeper interface {
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	//LockedCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	GetSupply(ctx context.Context, denom string) sdk.Coin
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	BlockedAddr(addr sdk.AccAddress) bool
	//BurnCoins(ctx context.Context, name string, amt sdk.Coins) error
	//UndelegateCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	//DelegateCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

type DistributionKeeper interface {
	GetValidatorDelegations(ctx context.Context, valAddr sdk.ValAddress) (delegations []stakingtypes.Delegation, err error)
	GetAllValidators(ctx context.Context) ([]stakingtypes.Validator, error)
}

type RewardKeeper interface {
	GetParams(ctx sdk.Context) (params rewardtypes.Params)
	GetValidatorSelfBondedTokens(ctx context.Context, val stakingtypes.ValidatorI) (sdkmath.LegacyDec, error)
	CalculateTotalSelfBondedTokens(ctx context.Context, validators []stakingtypes.Validator) (sdkmath.Int, error)
	CalculateFeeShare(amount sdkmath.LegacyDec, selfBonded sdkmath.LegacyDec, totalSelfBonded sdkmath.Int) sdkmath.LegacyDec
	CalculateMinApr(ctx context.Context, selfBonded sdkmath.LegacyDec) (seldbondtoken sdkmath.LegacyDec, err error)
}

type AllocationPoolKeeper interface {
	GetEmissionState(ctx context.Context) (val allocationpooltypes.Emission, found bool)
	SetEmission(ctx context.Context, emission allocationpooltypes.Emission) error
	GetDailyEmissionAmount(ctx context.Context) (val allocationpooltypes.CRADailyEmission, found bool)
	SetDailyEmission(ctx context.Context, emission allocationpooltypes.CRADailyEmission) error
	GetEmissionPerBlock(goCtx context.Context, blockNumber int64) (val string, found bool)
	GetEmissionEntityPerBlock(ctx context.Context, blockNumber string) (emission allocationpooltypes.Emission, found bool)
}

type MintKeeper interface {
	GetTotalMinted(ctx context.Context) (val gemminttypes.Minted, found bool)
}

type CRAKeeper interface {
	GetCRAAll(ctx context.Context) (list []cratypes.CRA)
	GetWalletAll(ctx context.Context) (list []cratypes.Wallet)
	GetWalletByWalletAddress(ctx context.Context, address string) (wallte cratypes.Wallet, found bool)
	GetCRAByCRAAddress(ctx context.Context, address string) (cra cratypes.CRA, found bool)
	SetCRA(ctx context.Context, cra cratypes.CRA) error
	SetWallet(ctx context.Context, cra cratypes.Wallet) error
}
