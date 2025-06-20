package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	customLog "cosmossdk.io/log"

	"strconv"

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

// Keeper - reward module keeper
func (k Keeper) DistributeRewards(ctx sdk.Context, validatorAddr sdk.ValAddress, transactionFees sdk.Coins) error {
	params := k.GetParams(ctx)

	validator, err := k.stakingKeeper.Validator(ctx, validatorAddr)
	if err != nil {
		return nil
	}

	// Skip if validator is jailed or not bonded
	if validator.IsJailed() || !validator.IsBonded() {
		return nil
	}

	selfDel, err := k.stakingKeeper.Delegation(ctx, sdk.AccAddress(validatorAddr), validatorAddr)
	fmt.Println("000000000000000000000000", selfDel)
	if err != nil {
		return err
	}

	totalSelfBonded := k.GetTotalSelfBondedTokens(ctx, validatorAddr)
	if totalSelfBonded.IsZero() {
		return nil
	}

	rewardCoins := sdk.NewCoins()
	emissionCoins := sdk.NewCoins()

	for _, fee := range transactionFees {
		// validator_fee_share = tx_fees × (self_bond / total_self_bond)
		feeShare := fee.Amount.Mul(sdkmath.Int(selfDel.GetShares())).Quo(totalSelfBonded)

		paramapr, err := strconv.ParseInt(params.Apr, 10, 64)
		if err != nil {
			return sdkerrors.Wrapf(errortypes.ErrInvalidType, "[UpdateAllocationPool][ParseInt] failed. Couldn't convert paramapr to int64. Error: [ %T ]", err)
		}

		paramblocksnumber, err := strconv.ParseInt(params.BlockPerYear, 10, 64)
		if err != nil {
			return sdkerrors.Wrapf(errortypes.ErrInvalidType, "[UpdateAllocationPool][ParseInt] failed. Couldn't convert paramblocksnumber to int64. Error: [ %T ]", err)
		}

		// minimum_APR_reward = self_bond × (APR / blocks_per_year)
		aprDecimal := sdkmath.LegacyNewDecWithPrec(paramapr, 2) // e.g., 4% = 0.04
		minAprReward := aprDecimal.MulInt(sdkmath.Int(selfDel.GetShares())).QuoInt64(paramblocksnumber)

		// Compare and determine reward
		if feeShare.GTE(minAprReward.TruncateInt()) {
			rewardCoins = rewardCoins.Add(sdk.NewCoin(fee.Denom, feeShare))
		} else {
			emissionNeeded := minAprReward.Sub(sdkmath.LegacyNewDecFromInt(feeShare)).TruncateInt()
			if emissionNeeded.IsPositive() {
				emission := sdk.NewCoin(fee.Denom, emissionNeeded)
				reward := sdk.NewCoin(fee.Denom, minAprReward.TruncateInt())
				rewardCoins = rewardCoins.Add(reward)
				emissionCoins = emissionCoins.Add(emission)
			}
		}
	}

	return nil
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

// func (k Keeper) Distribution(ctx context.Context) error {
// 	total := sdkmath.ZeroInt()

// 	k.stakingKeeper.IterateValidators(ctx, func(_ int64, val stakingtypes.ValidatorI) (stop bool) {
// 		fmt.Println("5555555555555555555555555")
// 		validator, _ := k.stakingKeeper.Validator(ctx, sdk.ValAddress(val.GetOperator()))
// 		fmt.Println("6666666666666666666666666666666")
// 		if validator != nil {
// 			total = total.Add(validator.GetBondedTokens())
// 		}
// 		return false
// 	})

// 	fmt.Println("total", total)

// 	k.stakingKeeper.IterateValidators(ctx, func(_ int64, val stakingtypes.ValidatorI) (stop bool) {
// 		validator, _ := k.stakingKeeper.Validator(ctx, sdk.ValAddress(val.GetOperator()))
// 		if validator != nil {
// 			intfrombool, _ := sdkmath.NewIntFromString("10")
// 			validatorFeeShare := intfrombool.Mul(validator.GetBondedTokens().Quo(total))
// 			fmt.Println("00000000000000validatorFeeShare000", validatorFeeShare)
// 		}
// 		return false
// 	})

// 	return nil
// }

// func ConvertAddrsToValAddrs(addrs []sdk.AccAddress) []sdk.ValAddress {
// 	valAddrs := make([]sdk.ValAddress, len(addrs))

// 	for i, addr := range addrs {
// 		valAddrs[i] = sdk.ValAddress(addr)
// 	}

// 	return valAddrs
// }

// CalculateRewards calculates and distributes validator rewards for a block
// func (rm Keeper) CalculateRewards(ctx sdk.Context, blockFees sdk.Coins) error {
// 	params := rm.GetParams(ctx)
// 	// Get all validators
// 	validators, _ := rm.stakingKeeper.GetAllValidators(ctx)

// 	// Calculate total self-bonded tokens
// 	totalSelfBonded := sdkmath.NewInt(0)
// 	for _, val := range validators {
// 		validator, _ := rm.stakingKeeper.Validator(ctx, sdk.ValAddress(val.GetOperator()))
// 		totalSelfBonded = totalSelfBonded.Add(validator.GetTokens())
// 	}

// 	if totalSelfBonded.IsZero() {
// 		return nil // No validators to reward
// 	}

// 	// Process rewards for each validator
// 	for _, validator := range validators {
// 		valAddr := validator.GetOperator()
// 		validator, _ := rm.stakingKeeper.Validator(ctx, sdk.ValAddress(valAddr))
// 		//selfBonded := rm.stakingKeeper.GetValidator(ctx, valAddr).GetTokens()

// 		blockFees := sdk.NewCoin("outbe", sdkmath.NewInt(10))
// 		// Calculate validator's fee share
// 		validatorFeeShare := sdkmath.LegacyNewDecFromInt(blockFees.Amount).
// 			Mul(sdkmath.LegacyNewDecFromInt(validator.GetTokens())).
// 			Quo(sdkmath.LegacyNewDecFromInt(totalSelfBonded))

// 		// Calculate minimum APR reward
// 		minAPR := sdkmath.LegacyMustNewDecFromStr(params.Apr).Quo(sdkmath.LegacyMustNewDecFromStr(params.BlockPerYear))
// 		minAPRReward := sdkmath.LegacyNewDecFromInt(validator.GetTokens()).Mul(minAPR)

// 		reward := sdk.NewCoin("outbe", validatorFeeShare.TruncateInt())
// 		emission := sdk.NewCoin("outbe", sdkmath.NewInt(0))

// 		// Check if additional emission is needed
// 		if validatorFeeShare.LT(minAPRReward) {
// 			requiredReward := minAPRReward.TruncateInt()
// 			emissionAmount := requiredReward.Sub(validatorFeeShare.TruncateInt())
// 			emission = sdk.NewCoin("outbe", emissionAmount)
// 			reward = sdk.NewCoin("outbe", requiredReward)
// 		}

// 		totalEmmition, _ := rm.allocationPoolKeeper.GetTotalEmission(ctx)
// 		poolEmission, _ := strconv.ParseUint(totalEmmition.TotalEmission, 10, 64)
// 		// Verify sufficient allocation pool
// 		if emission.Amount.GT(sdkmath.NewIntFromUint64(poolEmission)) {
// 			return sdkerrors.Wrap(errortypes.ErrInvalidType, "allocation pool depleted")
// 		}

// 		// Distribute rewards
// 		if !reward.Amount.IsZero() {
// 			err := rm.bankKeeper.SendCoinsFromModuleToAccount(
// 				ctx,
// 				types.ModuleName,
// 				sdk.AccAddress(valAddr),
// 				sdk.NewCoins(reward),
// 			)
// 			if err != nil {
// 				return err
// 			}
// 		}

// 		// Mint emission if needed
// 		if !emission.Amount.IsZero() {
// 			err := rm.mintKeeper.MintCoins(ctx, sdk.NewCoins(emission))
// 			if err != nil {
// 				return err
// 			}
// 			rm.allocationPool = rm.allocationPool.Sub(sdk.NewCoins(emission))

// 			err = rm.bankKeeper.SendCoinsFromModuleToAccount(
// 				ctx,
// 				mint.ModuleName,
// 				sdk.AccAddress(valAddr),
// 				sdk.NewCoins(emission),
// 			)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 	}

// 	return nil
// }

// UpdateAllocationPool updates the allocation pool based on TBD formula
// func (rm *RewardModule) UpdateAllocationPool(ctx sdk.Context) {
// 	// TBD: Implement allocation pool update logic
// 	// This could be based on network parameters, block fill ratios, etc.
// }

// // BeginBlock executes reward calculation at the start of each block
// func (rm *RewardModule) BeginBlock(ctx sdk.Context, blockFees sdk.Coins) {
// 	if err := rm.CalculateRewards(ctx, blockFees); err != nil {
// 		ctx.Logger().Error("Failed to calculate rewards", "error", err)
// 	}
// 	rm.UpdateAllocationPool(ctx)
// }

// GenesisState defines the reward module's genesis state
// type GenesisState struct {
// 	AllocationPool sdk.Coins `json:"allocation_pool"`
// 	APR            sdk.Dec   `json:"apr"`
// 	MaxStake       sdk.Int   `json:"max_stake"`
// }

// InitGenesis initializes the reward module's state from a provided genesis state
// func (rm *RewardModule) InitGenesis(ctx sdk.Context, data json.RawMessage) error {
// 	var genState GenesisState
// 	if err := json.Unmarshal(data, &genState); err != nil {
// 		return err
// 	}

// 	rm.allocationPool = genState.AllocationPool
// 	if genState.APR.IsZero() {
// 		genState.APR = sdk.NewDecFromFloat(AnnualPercentageRate)
// 	}
// 	// Validate max stake
// 	if genState.MaxStake.IsZero() {
// 		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "max stake not set")
// 	}

// 	return nil
// }

// ExportGenesis exports the reward module's genesis state
// func (rm *RewardModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
// 	genState := GenesisState{
// 		AllocationPool: rm.allocationPool,
// 		APR:            sdk.NewDecFromFloat(AnnualPercentageRate),
// 		MaxStake:       rm.stakingKeeper.GetParams(ctx).MaxStake,
// 	}

// 	data, _ := json.Marshal(genState)
// 	return data
// }

// validator_fee_share = transaction_fees_in_block × (validator_self_bonded_tokens / total_self_bonded_tokens)
// func (k Keeper) CalculateValidatorReward(ctx sdk.Context, selfbond sdkmath.LegacyDec, totalStaked sdkmath.Int) sdkmath.Int {
// 	return sdkmath.NewInt(10).Mul(selfbond.Quo(totalStaked))
// }

// // minimum_APR_reward = validator_self_bonded_tokens × (APR / blocks_per_year)
// func (k Keeper) CalculateMinRewardAmount(ctx sdk.Context, selfbond sdkmath.Int) sdkmath.Int {
// 	params := k.GetParams(ctx)

// 	apr, _ := sdkmath.NewIntFromString(params.Apr)
// 	blockPerYear, _ := sdkmath.NewIntFromString(params.BlockPerYear)

// 	return selfbond.Mul(apr.Quo(blockPerYear))
// }

// ------------------------------------------------------------------------

// Transaction fees collected within the block.
// Additional token emission if transaction fees are insufficient.
// Transaction fees are distributed among validators proportionally based on their self-bonded tokens.
// The maximum total stake (limit on self-bonded tokens) is specified in the genesis configuration (see TBD).

// Reward Calculation Formula

// validator_fee_share = transaction_fees_in_block × (validator_self_bonded_tokens / total_self_bonded_tokens)

// minimum_APR_reward = validator_self_bonded_tokens × (APR / blocks_per_year)  // 'blocks_per_year' is calculated based on the expected average block time; e.g., with 5-second blocks, this results in approximately 6,307,200 blocks annually.

// if validator_fee_share ? minimum_APR_reward:
//     reward = validator_fee_share
//     emission = 0
// else:
//     reward = minimum_APR_reward
//     emission = minimum_APR_reward - validator_fee_share

// Emission tokens are minted only if the validator's fee share does not meet their minimum APR requirement.

// Example #1 Calculation:
// Validator A has 1,000 self-bonded tokens.
// Total self-bonded tokens in the network = 100,000.
// APR = 4% (is permanently set in the genesis configuration)
// Blocks per year = 6,307,200 (assuming 5-second blocks).

// Transaction fees in block = 10 tokens.

// validator_fee_share = 10 × (1,000 / 100,000) = 0.1 tokens
// minimum_APR_reward = 1,000 × (0.04 / 6,307,200) ? 0.00000634 tokens

// Validator A receives 0.1 tokens from fees, and no additional emission is required since the fee share exceeds the minimum APR reward.

// Example #2 Calculation:

// Validator B has 10,000 self-bonded tokens.
// Total self-bonded tokens in the network = 100,000.
// APR = 4% (is permanently set in the genesis configuration)
// Blocks per year = 6,307,200 (assuming 5-second blocks).
// Transaction fees in block = 1 tokens.

// validator_fee_share = 0.01 × (10,000 / 100,000) = 0.1 tokens
// minimum_APR_reward = 10,000 × (0.04 / 6,307,200) ? 0.634 tokens
// emission = 0.634 - 0.1 = 0.534 (minimum_APR_reward - validator_fee_share)

// Validator B receives 0.1 tokens from fees, and additional emission 0.534 is required since the fee share less the minimum APR reward.
