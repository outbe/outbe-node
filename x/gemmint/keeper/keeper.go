package keeper

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/log"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/outbe/outbe-node/app/params"
	"github.com/outbe/outbe-node/x/gemmint/types"
)

// Keeper of the mint store
type Keeper struct {
	cdc              codec.BinaryCodec
	storeService     storetypes.KVStoreService
	stakingKeeper    types.StakingKeeper
	accountKeeper    types.AccountKeeper
	bankKeeper       types.BankKeeper
	rewardKeeper     types.RewardKeeper
	feeCollectorName string

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string

	Schema collections.Schema
	Params collections.Item[types.Params]
	Minter collections.Item[types.Minter]
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService storetypes.KVStoreService,
	sk types.StakingKeeper,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	rk types.RewardKeeper,
	feeCollectorName string,
	authority string,
) Keeper {
	// ensure mint module account is set
	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("the x/%s module account has not been set", types.ModuleName))
	}

	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		cdc:              cdc,
		storeService:     storeService,
		stakingKeeper:    sk,
		accountKeeper:    ak,
		bankKeeper:       bk,
		rewardKeeper:     rk,
		feeCollectorName: feeCollectorName,
		authority:        authority,
		Params:           collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		Minter:           collections.NewItem(sb, types.MinterKey, "minter", codec.CollValue[types.Minter](cdc)),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

func (k Keeper) GetAuthority() string {
	return k.authority
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	logger := ctx.Logger().With(
		"height", ctx.BlockHeight(),
		"timestamp", time.Now().UTC().Format(time.RFC3339),
		"module", fmt.Sprintf("x/%s", types.ModuleName),
	)
	return logger
}

func (k Keeper) StakingTokenSupply(ctx context.Context) (sdkmath.Int, error) {
	tokenSupply := k.bankKeeper.GetSupply(ctx, params.BondDenom).Amount
	return tokenSupply, nil
}

func (k Keeper) BondedRatio(ctx context.Context) (sdkmath.LegacyDec, error) {
	return k.stakingKeeper.BondedRatio(ctx)
}

func (k Keeper) MintCoins(ctx context.Context, newCoins sdk.Coins) error {
	if newCoins.Empty() {
		// skip as no coins need to be minted
		return nil
	}

	return k.bankKeeper.MintCoins(ctx, types.ModuleName, newCoins)
}

func (k Keeper) AddCollectedFees(ctx context.Context, fees sdk.Coins) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	k.Logger(sdkCtx).Info("[AddCollectedFees] fetching fee amount", "fees", fees)

	return k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, k.feeCollectorName, fees)
}
