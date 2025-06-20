package types

import (
	context "context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	allocationpooltypes "github.com/outbe/outbe-node/x/allocationpool/types"
	gemminttypes "github.com/outbe/outbe-node/x/gemmint/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
	HasAccount(ctx context.Context, addr sdk.AccAddress) bool
	GetModuleAddress(name string) sdk.AccAddress
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
}

type StakingKeeper interface {
	StakingTokenSupply(ctx context.Context) (math.Int, error)
	BondedRatio(ctx context.Context) (math.LegacyDec, error)
	GetAllValidators(ctx context.Context) (validators []stakingtypes.Validator, err error)
	GetValidatorDelegations(ctx context.Context, valAddr sdk.ValAddress) (delegations []stakingtypes.Delegation, err error)
	IterateValidators(ctx context.Context, fn func(index int64, validator stakingtypes.ValidatorI) (stop bool)) error
	Delegation(context.Context, sdk.AccAddress, sdk.ValAddress) (stakingtypes.DelegationI, error)
	Validator(context.Context, sdk.ValAddress) (stakingtypes.ValidatorI, error)
}

type MintKeeper interface {
	GetWhitelist(ctx context.Context) (list []gemminttypes.Whitelist)
	IsEligibleSmartContract(ctx context.Context, contractAddress string) bool
	SetTotalMinted(ctx context.Context, totalMinted gemminttypes.Minted) error
	GetTotalMinted(ctx context.Context) (totalMinted gemminttypes.Minted, found bool)
}

type AllocatioPoolKeeper interface {
	GetTotalEmission(ctx context.Context) (val allocationpooltypes.Emission, found bool)
}
