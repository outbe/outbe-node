package types

import (
	"context"

	"cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	rewardtypes "github.com/outbe/outbe-node/x/reward/types"
)

// AccountKeeper defines the account contract that must be fulfilled when
// creating a x/bank keeper.
type AccountKeeper interface {
	AddressCodec() address.Codec

	NewAccount(context.Context, sdk.AccountI) sdk.AccountI
	NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI

	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	GetAllAccounts(ctx context.Context) []sdk.AccountI
	HasAccount(ctx context.Context, addr sdk.AccAddress) bool
	SetAccount(ctx context.Context, acc sdk.AccountI)

	IterateAccounts(ctx context.Context, process func(sdk.AccountI) bool)

	ValidatePermissions(macc sdk.ModuleAccountI) error

	GetModuleAddress(moduleName string) sdk.AccAddress
	GetModuleAddressAndPermissions(moduleName string) (addr sdk.AccAddress, permissions []string)
	GetModuleAccountAndPermissions(ctx context.Context, moduleName string) (sdk.ModuleAccountI, []string)
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
	SetModuleAccount(ctx context.Context, macc sdk.ModuleAccountI)
	GetModulePermissions() map[string]types.PermissionsForAddress
}

type StakingKeeper interface {
	GetValidatorDelegations(ctx context.Context, valAddr sdk.ValAddress) (delegations []stakingtypes.Delegation, err error)
	GetAllValidators(ctx context.Context) ([]stakingtypes.Validator, error)
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
