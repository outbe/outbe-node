package keeper

import (
	"context"
	"fmt"
	"strconv"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	customLog "cosmossdk.io/log"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	errortypes "github.com/outbe/outbe-node/errors"
	"github.com/outbe/outbe-node/x/reward/types"
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

// calculateTotalSelfBondedTokens sums the self-bonded tokens of all validators in bondedVotes.
func (k Keeper) CalculateTotalSelfBondedTokens(ctx context.Context, bondedVotes []abci.VoteInfo) (sdkmath.Int, error) {
	total := sdkmath.ZeroInt()
	for _, vote := range bondedVotes {
		validator, err := k.stakingKeeper.ValidatorByConsAddr(ctx, vote.Validator.Address)
		if err != nil {
			return sdkmath.Int{}, sdkerrors.Wrapf(errortypes.ErrNoValidatorForAddress, "invalid validatir cons adddress")
		}
		selfBonded, err := k.GetValidatorSelfBondedTokens(ctx, validator)
		if err != nil {
			return sdkmath.Int{}, sdkerrors.Wrapf(errortypes.ErrInvalidRequest, "couldn't fetch validator self bond token")
		}
		total = total.Add(selfBonded.TruncateInt())
	}
	return total, nil
}

// getValidatorSelfBondedTokens retrieves the self-bonded tokens for a validator.
func (k Keeper) GetValidatorSelfBondedTokens(ctx context.Context, val stakingtypes.ValidatorI) (sdkmath.LegacyDec, error) {
	valBz, err := k.stakingKeeper.ValidatorAddressCodec().StringToBytes(val.GetOperator())
	if err != nil {
		return sdkmath.LegacyDec{}, sdkerrors.Wrapf(errortypes.ErrInvalidRequest, "failed to convert validator address to bytes: %s", err.Error())
	}

	valAddr := sdk.ValAddress(valBz)
	delAddr := sdk.AccAddress(valBz)

	selfDel, err := k.stakingKeeper.Delegation(ctx, delAddr, valAddr)
	if err != nil {
		return sdkmath.LegacyDec{}, sdkerrors.Wrapf(errortypes.ErrInvalidRequest, "failed to query delegation for delegator %s and validator %s: %s", delAddr.String(), valAddr.String(), err.Error())
	}

	tokens := val.TokensFromShares(selfDel.GetShares())
	return tokens, nil
}

func (k Keeper) CalculateFeeShare(amount sdkmath.LegacyDec, selfBonded sdkmath.LegacyDec, totalSelfBonded sdkmath.Int) sdkmath.LegacyDec {
	return amount.Mul(selfBonded).QuoInt(totalSelfBonded)
}

func (k Keeper) CalculateMinApr(ctx context.Context, selfBonded sdkmath.LegacyDec) sdkmath.LegacyDec {
	sdkctx := sdk.UnwrapSDKContext(ctx)
	params := k.GetParams(sdkctx)

	num, err := strconv.ParseInt(params.BlockPerYear, 10, 64)
	if err != nil {
		fmt.Printf("Error converting string to int64: %v\n", err)
		return sdkmath.LegacyDec{}
	}

	minAprFactor := sdkmath.LegacyMustNewDecFromStr(params.Apr).QuoInt64(num)
	return selfBonded.Mul(minAprFactor)
}
