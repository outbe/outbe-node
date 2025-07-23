package keeper

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"sort"
	"time"

	"cosmossdk.io/core/store"
	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/outbe/outbe-node/errors"
	"github.com/outbe/outbe-node/x/rand/types"
)

var (
	PrintLoges bool
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeService store.KVStoreService

		stakingKeeper    types.StakingKeeper
		accountKeeper    types.AccountKeeper
		bankKeeper       types.BankKeeper
		feeCollectorName string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,

	stakingKeeper types.StakingKeeper,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	feeCollectorName string,
) Keeper {

	return Keeper{

		cdc:          cdc,
		storeService: storeService,

		stakingKeeper:    stakingKeeper,
		accountKeeper:    accountKeeper,
		bankKeeper:       bankKeeper,
		feeCollectorName: feeCollectorName,
	}
}

// GetLogger returns a logger instance with optional log printing based on the PrintLogs environment variable.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	logger := ctx.Logger().With(
		"timestamp", time.Now().UTC().Format(time.RFC3339),
		"module", fmt.Sprintf("x/%s", types.ModuleName),
		"height", ctx.BlockHeight(),
	)
	return logger
}

func (k Keeper) GetPeriod(ctx context.Context) (val types.Period, found bool) {
	store := k.storeService.OpenKVStore(ctx)
	periodKey := types.GetPeriodKey("period")
	b, err := store.Get(periodKey)

	if b == nil || err != nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

func (k Keeper) SetPeriod(ctx context.Context, period types.Period) error {
	store := k.storeService.OpenKVStore(ctx)
	b := k.cdc.MustMarshal(&period)
	return store.Set(types.GetPeriodKey("period"), b)
}

func (k Keeper) SetCommitment(ctx sdk.Context, commitment types.Commitment) {
	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&commitment)
	store.Set(types.GetCommitmentKey(commitment.Period, commitment.Validator), bz)
}

func (k Keeper) GetCommitment(ctx context.Context, period uint64, validator string) (val types.Commitment, found bool) {

	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.GetCommitmentKey(period, validator))

	if bz == nil || err != nil {
		return val, false
	}

	k.cdc.MustUnmarshal(bz, &val)
	return val, true
}

func (k Keeper) GetCommitments(ctx context.Context) []*types.Commitment {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iterator := storetypes.KVStorePrefixIterator(store, types.CommitmentKey)

	defer iterator.Close()

	var commitments []*types.Commitment
	for ; iterator.Valid(); iterator.Next() {
		var val types.Commitment
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		commitments = append(commitments, &val)
	}
	return commitments
}

func (k Keeper) GetCommitmentsByPeriod(ctx context.Context, period uint64) (list []types.Commitment) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iterator := storetypes.KVStorePrefixIterator(store, types.CommitmentKey)

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Commitment
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		if val.Period == period {
			list = append(list, val)
		}
	}
	return
}

func (k Keeper) HasCommitment(ctx sdk.Context, period uint64, validator string) bool {
	store := k.storeService.OpenKVStore(ctx)
	prefixStore := prefix.NewStore(runtime.KVStoreAdapter(store), types.CommitmentKey)
	return prefixStore.Has(types.GetCommitmentKey(period, validator))
}

func (k Keeper) SetReveal(ctx context.Context, reveal types.Reveal) error {
	store := k.storeService.OpenKVStore(ctx)
	valAddress, _ := sdk.ValAddressFromBech32(reveal.Validator)
	bz := k.cdc.MustMarshal(&reveal)
	return store.Set(types.GetRevealKey(reveal.Period, valAddress), bz)
}

func (k Keeper) SetPenalty(ctx sdk.Context, penalty types.Penalty) {
	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&penalty)
	store.Set(types.GetPenaltyKey(penalty.Period, sdk.ValAddress(penalty.Validator)), bz)
}

func (k Keeper) GetPenalty(ctx context.Context, period uint64, validator string) (val types.Penalty, found bool) {

	store := k.storeService.OpenKVStore(ctx)

	bz, err := store.Get(types.GetPenaltyKey(period, sdk.ValAddress(validator)))
	if bz == nil || err != nil {
		return val, false
	}

	k.cdc.MustUnmarshal(bz, &val)
	return val, true
}

func (k Keeper) GetPenalties(ctx context.Context) []*types.Penalty {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iterator := storetypes.KVStorePrefixIterator(store, types.PenaltyKey)

	defer iterator.Close()

	var penalties []*types.Penalty
	for ; iterator.Valid(); iterator.Next() {
		var val types.Penalty
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		penalties = append(penalties, &val)
	}
	return penalties
}

func (k Keeper) GetReveals(ctx context.Context) []*types.Reveal {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iterator := storetypes.KVStorePrefixIterator(store, types.RevealKey)

	defer iterator.Close()

	var reveals []*types.Reveal
	for ; iterator.Valid(); iterator.Next() {
		var val types.Reveal
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		reveals = append(reveals, &val)
	}
	return reveals
}

func (k Keeper) GetRevealsForPeriod(ctx context.Context, period uint64) []*types.Reveal {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iterator := storetypes.KVStorePrefixIterator(store, types.RevealKey)

	defer iterator.Close()

	var reveals []*types.Reveal
	for ; iterator.Valid(); iterator.Next() {
		var val types.Reveal
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		reveals = append(reveals, &val)
	}
	return reveals
}

func (k Keeper) GetCommitmentsForPeriod(ctx sdk.Context, period uint64) []*types.Commitment {

	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iterator := storetypes.KVStorePrefixIterator(store, sdk.Uint64ToBigEndian(period))

	defer iterator.Close()

	var commitments []*types.Commitment
	for ; iterator.Valid(); iterator.Next() {
		var commitment types.Commitment
		k.cdc.MustUnmarshal(iterator.Value(), &commitment)
		commitments = append(commitments, &commitment)
	}
	return commitments
}

func (k Keeper) GenerateRandomness(ctx sdk.Context, period uint64) ([]byte, error) {
	logger := k.Logger(ctx)

	// Get reveals for the period
	reveals := k.GetRevealsForPeriod(ctx, period)
	if len(reveals) == 0 {
		logger.Info("no reveals found, using block header app hash",
			"period", period)
		return ctx.BlockHeader().AppHash, nil
	}

	// Sort reveals by validator address
	sort.Slice(reveals, func(i, j int) bool {
		valAddressI, err := sdk.ValAddressFromBech32(reveals[i].Validator)
		if err != nil {
			logger.Error("failed to parse validator address for sorting",
				"validator", reveals[i].Validator,
				"period", period,
				"error", err)
			return false
		}
		valAddressJ, err := sdk.ValAddressFromBech32(reveals[j].Validator)
		if err != nil {
			logger.Error("failed to parse validator address for sorting",
				"validator", reveals[j].Validator,
				"period", period,
				"error", err)
			return false
		}
		return bytes.Compare(valAddressI.Bytes(), valAddressJ.Bytes()) < 0
	})

	// Combine reveal values
	combinedValue := []byte{}
	for _, reveal := range reveals {
		if reveal.RevealValue == nil {
			logger.Error("nil reveal value encountered",
				"validator", reveal.Validator,
				"period", period)
			return nil, sdkerrors.Wrap(errortypes.ErrInvalidReveal, "nil reveal value")
		}
		combinedValue = append(combinedValue, reveal.RevealValue...)
	}

	// Add block hash
	if ctx.BlockHeader().AppHash == nil {
		logger.Error("nil block header app hash",
			"period", period)
		return nil, sdkerrors.Wrap(errortypes.ErrInvalidState, "nil block header app hash")
	}
	combinedValue = append(combinedValue, ctx.BlockHeader().AppHash...)

	// Compute final hash
	hash := sha256.Sum256(combinedValue)
	logger.Info("randomness generated successfully",
		"period", period,
		"reveal_count", len(reveals))

	return hash[:], nil
}

func (k Keeper) PenalizeNonRevealers(ctx sdk.Context, period uint64) error {
	// logger := k.Logger(ctx)

	// validators, err := k.stakingKeeper.GetAllValidators(ctx)
	// if err != nil {
	// 	logger.Error("Failed to get validators",
	// 		"error", err,
	// 	)
	// 	return sdkerrors.Wrapf(errortypes.ErrInvalidPhase, "failed to get validators: %s", err)
	// }

	// if len(validators) == 1 {
	// 	return nil
	// }

	commitments := k.GetCommitmentsByPeriod(ctx, period)
	for _, commitment := range commitments {
		if !commitment.Revealed {
			valAddress, _ := sdk.ValAddressFromBech32(commitment.Validator)
			validator, _ := k.stakingKeeper.Validator(ctx, valAddress)

			if validator != nil && validator.IsBonded() {

				// Deposit is not returned (stays in module)
				penalty := types.Penalty{
					Period:    commitment.Period,
					Validator: commitment.Validator,
					Deposit:   commitment.Deposit,
				}
				k.SetPenalty(ctx, penalty)

				// Optional: additional penalty through slashing module [Ignored]

				// k.stakingKeeper.Slash(
				// 	ctx,
				// 	conAddress,
				// 	ctx.BlockHeight(),
				// 	validator.GetConsensusPower(sdk.DefaultPowerReduction),
				// 	math.LegacyNewDec(params.PenaltyAmount.Amount.Int64()),
				// )

				// params := k.GetParams(ctx)
				// conAddress, _ := validator.GetConsAddr()

				ctx.EventManager().EmitEvent(
					sdk.NewEvent(
						types.EventTypePenalty,
						sdk.NewAttribute(types.AttributeKeyValidator, commitment.Validator),
						sdk.NewAttribute(types.AttributeKeyPeriodNumber, fmt.Sprintf("%d", period)),
					),
				)
			}
		}
	}
	return nil
}

func (k Keeper) ClearPeriodData(ctx context.Context, period uint64) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	commitmentStore := prefix.NewStore(store, types.CommitmentKey)
	iterator := commitmentStore.Iterator(nil, nil)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		store.Delete(iterator.Key())
	}

	store = runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	revealStore := prefix.NewStore(store, types.RevealKey)
	iterator = revealStore.Iterator(nil, nil)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		store.Delete(iterator.Key())
	}
}
