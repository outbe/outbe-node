package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	customLog "cosmossdk.io/log"

	sdkerrors "cosmossdk.io/errors"
	errortypes "github.com/outbe/outbe-node/errors"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/outbe/outbe-node/x/reward/types"
)

// Keeper of the mint store
type Keeper struct {
	cdc                  codec.BinaryCodec
	storeService         storetypes.KVStoreService
	stakingKeeper        types.StakingKeeper
	bankKeeper           types.BankKeeper
	allocationPoolKeeper types.AllocatioPoolKeeper
	feeCollectorName     string

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string

	Schema collections.Schema
	Params collections.Item[types.Params]
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService storetypes.KVStoreService,
	sk types.StakingKeeper,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	alp types.AllocatioPoolKeeper,
	feeCollectorName string,
	authority string,
) Keeper {
	// ensure mint module account is set
	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("the x/%s module account has not been set", types.ModuleName))
	}

	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		cdc:                  cdc,
		storeService:         storeService,
		stakingKeeper:        sk,
		bankKeeper:           bk,
		allocationPoolKeeper: alp,
		feeCollectorName:     feeCollectorName,
		authority:            authority,
		Params:               collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
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

func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &params)
	return
}

func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := k.cdc.MustMarshal(&params)
	store.Set(types.ParamsKey, bz)
	return nil
}

// UpdateAllocationPool updates the allocation pool based on TBD formula
func (k Keeper) UpdateAllocationPool(ctx sdk.Context) {
	// TBD: Implement allocation pool update logic
	// This could be based on network parameters, block fill ratios, etc.
}

func (k Keeper) GetTotalSelfBondedTokens(ctx sdk.Context, validatorAddr sdk.ValAddress) sdkmath.Int {
	total := sdkmath.ZeroInt()
	k.stakingKeeper.IterateValidators(ctx, func(_ int64, val stakingtypes.ValidatorI) (stop bool) {
		selfDel, _ := k.stakingKeeper.Delegation(ctx, sdk.AccAddress(validatorAddr), validatorAddr)
		if selfDel != nil {
			total = total.Add(selfDel.GetShares().TruncateInt())
		}
		return false
	})
	return total
}

func (k Keeper) CalculateValidatorFeeShare(transactionFees, selfBondedTokens, totalSelfBondedTokens sdkmath.LegacyDec) (sdkmath.LegacyDec, error) {
	// Check for zero or negative totalSelfBondedTokens to prevent division by zero
	if totalSelfBondedTokens.IsZero() {
		return sdkmath.LegacyZeroDec(), sdkerrors.Wrap(errortypes.ErrInvalidRequest, "total self-bonded tokens cannot be zero")
	}
	if totalSelfBondedTokens.IsNegative() {
		return sdkmath.LegacyZeroDec(), sdkerrors.Wrap(errortypes.ErrInvalidRequest, "total self-bonded tokens cannot be negative")
	}

	// Check for negative transactionFees or selfBondedTokens
	if transactionFees.IsNegative() {
		return sdkmath.LegacyZeroDec(), sdkerrors.Wrap(errortypes.ErrInvalidRequest, "transaction fees cannot be negative")
	}
	if selfBondedTokens.IsNegative() {
		return sdkmath.LegacyZeroDec(), sdkerrors.Wrap(errortypes.ErrInvalidRequest, "self-bonded tokens cannot be negative")
	}

	// validator_fee_share = transactionFees × (selfBondedTokens / totalSelfBondedTokens)
	shareRatio, err := sdkmath.LegacyNewDecFromStr((selfBondedTokens.Quo(totalSelfBondedTokens)).String())
	if err != nil {
		return sdkmath.LegacyZeroDec(), sdkerrors.Wrapf(errortypes.ErrInvalidRequest, "failed to calculate share ratio: %v", err)
	}

	validatorFeeShare, err := sdkmath.LegacyNewDecFromStr((transactionFees.Mul(shareRatio)).String())
	if err != nil {
		return sdkmath.LegacyZeroDec(), sdkerrors.Wrapf(errortypes.ErrInvalidRequest, "failed to calculate validator fee share: %v", err)
	}

	return validatorFeeShare, nil
}

// CalculateMinimumAPRReward calculates the minimum APR reward based on the self-bonded tokens
// and the annual percentage rate (APR).
func (k Keeper) CalculateMinimumAPRReward(selfBondedTokens, apr sdkmath.LegacyDec, blocksPerYear int64) (sdkmath.LegacyDec, error) {
	if selfBondedTokens.IsNegative() {
		return sdkmath.LegacyZeroDec(), sdkerrors.Wrap(errortypes.ErrInvalidRequest, "self-bonded tokens cannot be negative")
	}
	if selfBondedTokens.IsZero() {
		return sdkmath.LegacyZeroDec(), sdkerrors.Wrap(errortypes.ErrInvalidRequest, "self-bonded tokens cannot be zero")
	}

	if apr.IsNegative() {
		return sdkmath.LegacyZeroDec(), sdkerrors.Wrap(errortypes.ErrInvalidRequest, "APR cannot be negative")
	}
	if apr.IsZero() {
		return sdkmath.LegacyZeroDec(), sdkerrors.Wrap(errortypes.ErrInvalidRequest, "APR cannot be zero")
	}

	if blocksPerYear <= 0 {
		return sdkmath.LegacyZeroDec(), sdkerrors.Wrap(errortypes.ErrInvalidRequest, "blocks per year must be positive")
	}

	blocksPerYearDec := sdkmath.LegacyNewDec(blocksPerYear)
	aprPerBlock, err := sdkmath.LegacyNewDecFromStr((apr.Quo(blocksPerYearDec)).String())
	if err != nil {
		return sdkmath.LegacyZeroDec(), sdkerrors.Wrapf(errortypes.ErrInvalidRequest, "failed to calculate APR per block: %v", err)
	}

	result, err := sdkmath.LegacyNewDecFromStr((selfBondedTokens.Mul(aprPerBlock)).String())
	if err != nil {
		return sdkmath.LegacyZeroDec(), sdkerrors.Wrapf(errortypes.ErrInvalidRequest, "failed to calculate minimum APR reward: %v", err)
	}

	return result, nil
}
