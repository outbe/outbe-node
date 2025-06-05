package keeper

import (
	"context"
	"fmt"
	"log"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	customLog "cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/outbe/outbe-node/x/gemmint/constants"
	"github.com/outbe/outbe-node/x/gemmint/types"
)

// Keeper of the mint store
type Keeper struct {
	cdc              codec.BinaryCodec
	storeService     storetypes.KVStoreService
	stakingKeeper    types.StakingKeeper
	bankKeeper       types.BankKeeper
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
		bankKeeper:       bk,
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

func (k Keeper) Logger(ctx context.Context) customLog.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", "x/"+types.ModuleName)
}

func (k Keeper) StakingTokenSupply(ctx context.Context) (sdkmath.Int, error) {
	return k.stakingKeeper.StakingTokenSupply(ctx)
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

	k.Logger(ctx).Info("[AddCollectedFees] fetching fee amount", "fees", fees)

	return k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, k.feeCollectorName, fees)
}

const (
	BlocksPerYear = 6311520
	ValidatorAPR  = "0.4"
	Decimals      = 6
)

func (k Keeper) CalculateValidatorReward(ctx context.Context) {

	log.Println("########## Calculating Validator Reward is Started ##########")

	var circulating sdkmath.Int

	params, err := k.Params.Get(ctx)
	if err != nil {
		panic(err)
	}

	totalMinted, found := k.GetTotalMinted(ctx)
	if !found {
		circulating = sdkmath.NewInt(0)
	} else {
		circulating = totalMinted.TotalMinted.TruncateInt()
	}

	aprPerBlock := constants.ValidatorAPR.QuoInt64(int64(constants.BlocksPerYear))

	decimalScale := sdkmath.LegacyNewDec(1_000_000)
	scaledAprPerBlock := aprPerBlock.Mul(decimalScale)

	rewardAmount := scaledAprPerBlock.MulInt(circulating).TruncateInt()
	if rewardAmount.IsZero() {
		k.Logger(ctx).Info("[CalculateValidatorReward] Low Supply Amount. No Reward", "reward_amount", rewardAmount)
		return
	}

	rewardCoin := sdk.NewCoin(params.MintDenom, rewardAmount)
	coins := sdk.NewCoins(rewardCoin)

	k.Logger(ctx).Info("[CalculateValidatorReward] coins", "coins", coins)

	mintingError := k.MintCoins(ctx, coins)
	if mintingError != nil {
		panic(mintingError)
	}

	err = k.AddCollectedFees(ctx, coins)
	if err != nil {
		panic(err)
	}

	minted := types.Minted{
		TotalMinted: totalMinted.TotalMinted.Add(sdkmath.LegacyNewDecFromInt(coins[0].Amount)),
	}

	k.Logger(ctx).Info("[CalculateValidatorReward] minted", "minted", minted)

	totalmintError := k.SetTotalMinted(ctx, minted)
	if totalmintError != nil {
		panic(totalmintError)
	}
}
