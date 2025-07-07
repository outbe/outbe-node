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
	BlocksPerYear = "6307200"
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

// func (k Keeper) Calculate(ctx context.Context) error {

// 	totalSelfBonded := sdkmath.ZeroInt()
// 	validators, _ := k.stakingKeeper.GetAllValidators(ctx)
// 	for _, validator := range validators {
// 		if validator.IsJailed() || !validator.IsBonded() {
// 			return nil
// 		}
// 		totalSelfBonded = totalSelfBonded.Add(validator.GetTokens())
// 	}

// 	feeCollector := k.accountKeeper.GetModuleAccount(ctx, constants.FeeCollectorName)
// 	feesCollectedInt := k.bankKeeper.GetAllBalances(ctx, feeCollector.GetAddress())
// 	feesCollected := sdk.NewDecCoinsFromCoins(feesCollectedInt...)

// 	fmt.Println("666666666666666666666666666--feesCollected---before", feesCollected)

// 	if !feesCollectedInt.IsValid() {
// 		return sdkerrors.Wrapf(errortypes.ErrInvalidCoins, "invalid fees collected: %s", feesCollectedInt.String())
// 	}

// 	if feesCollectedInt.IsZero() {
// 		return nil
// 	}

// 	if constants.FeeCollectorName == "" || constants.ModuleName == "" {
// 		return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "module names cannot be empty")
// 	}

// 	totalMintTokens := sdkmath.LegacyZeroDec()

// 	for _, validator := range validators {

// 		fmt.Println("666666666666666666666666666--validator", validator)
// 		fmt.Println("666666666666666666666666666--sdkmath.LegacyNewDecFromInt(sdkmath.NewInt(1).Quo(sdkmath.NewInt(10)))", sdkmath.LegacyNewDecFromInt(sdkmath.NewInt(1).Quo(sdkmath.NewInt(10))))

// 		feeShare, err := k.rewardKeeper.CalculateValidatorFeeShare(
// 			sdkmath.LegacyNewDec(100),
// 			sdkmath.LegacyNewDec(10000),
// 			sdkmath.LegacyNewDec(100000),
// 		)
// 		if err != nil {
// 			return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "couldn't calculate validator fee share")
// 		}

// 		fmt.Println("666666666666666666666666666--feeShare", feeShare)

// 		i, err := strconv.ParseInt(BlocksPerYear, 10, 64)
// 		if err != nil {
// 			return fmt.Errorf("failed to convert string '%s' to int64: %v", ValidatorAPR, err)
// 		}
// 		fmt.Println("666666666666666666666666666--i", i)

// 		minAprreward, err := k.rewardKeeper.CalculateMinimumAPRReward(
// 			sdkmath.LegacyNewDec(10000),
// 			sdkmath.LegacyMustNewDecFromStr("0.04"),
// 			6307200,
// 		)
// 		fmt.Println("666666666666666666666666666--minAprreward", minAprreward)
// 		if err != nil {
// 			return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "failed to calculate min apr reward")
// 		}

// 		totalMintTokens = totalMintTokens.Add(feeShare)
// 		fmt.Println("totalMintTokens ---------after----feeShare--->", totalMintTokens)

// 		if feeShare.LTE(minAprreward) {
// 			emissionNeeded := minAprreward.Sub(feeShare)
// 			fmt.Println("minAprreward ---------------->", minAprreward)     // 0.000063419583970000
// 			fmt.Println("feeShare ---------------->", feeShare)             // 0.000000000000000000
// 			fmt.Println("emissionNeeded ---------------->", emissionNeeded) // 0.000063419583970000
// 			if emissionNeeded.IsNegative() && emissionNeeded.IsZero() {
// 				return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "not valid needed emission amount")
// 			}

// 			fmt.Println("here2 ---------------->")

// 			emission := sdk.NewDecCoinFromDec(constants.Denom, sdkmath.LegacyNewDecWithPrec(emissionNeeded.RoundInt64(), constants.LegacyPrecision))
// 			fmt.Println("emission ---------------->", emission)

// 			mintCoin := sdk.NewCoin(constants.Denom, minAprreward.RoundInt())
// 			mintCoins := sdk.NewCoins(mintCoin)
// 			fmt.Println("mintCoins ---------------->", mintCoins)

// 			if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, mintCoins); err != nil {
// 				return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[AllocateTokens][MintCoins] failed. Failed to mint coins")
// 			}

// 			totalMintTokens = totalMintTokens.Add(sdkmath.LegacyDec(mintCoins[0].Amount))
// 			fmt.Println("totalMintTokens ---------after----minapramount--->", totalMintTokens)
// 		}
// 	}

// 	totalFeeCollectedCoin := sdk.NewCoin(params.BondDenom, sdkmath.Int(totalMintTokens.TruncateInt()))

// 	minttotalFeeCollectedCoin := sdk.NewCoins(totalFeeCollectedCoin)
// 	fmt.Println("minttotalFeeCollectedCoin ---------------->", minttotalFeeCollectedCoin)

// 	//k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, constants.FeeCollectorName, minttotalFeeCollectedCoin)

// 	return nil
// }
