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
	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	appParams "github.com/outbe/outbe-node/app/params"
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
		if val.Period == period && !val.Revealed {
			list = append(list, val)
		}
	}
	return
}

func (k Keeper) GetValidatorCommitmentByPeriod(ctx context.Context, period uint64, validator string) (list []types.Commitment) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iterator := storetypes.KVStorePrefixIterator(store, types.CommitmentKey)

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Commitment
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		if val.Period == period && val.Validator == validator && !val.Revealed {
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
	logger := ctx.Logger().With("module", "keeper")

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
	logger := k.Logger(ctx)

	validators, err := k.stakingKeeper.GetAllValidators(ctx)
	if err != nil {
		logger.Error("Failed to get validators",
			"error", err,
		)
		return sdkerrors.Wrapf(errortypes.ErrInvalidPhase, "failed to get validators: %s", err)
	}

	if len(validators) == 1 {
		return nil
	}

	commitments := k.GetCommitmentsByPeriod(ctx, period)
	for _, commitment := range commitments {
		if !commitment.Revealed {
			valAddress, _ := sdk.ValAddressFromBech32(commitment.Validator)
			validator, _ := k.stakingKeeper.Validator(ctx, valAddress)

			if validator != nil && validator.IsBonded() {
				params := k.GetParams(ctx)
				conAddress, _ := validator.GetConsAddr()
				k.stakingKeeper.Slash(
					ctx,
					conAddress,
					ctx.BlockHeight(),
					validator.GetConsensusPower(sdk.DefaultPowerReduction),
					math.LegacyNewDec(params.PenaltyAmount.Amount.Int64()),
				)
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

func (k Keeper) ClearPeriodData(goCtx context.Context, period uint64) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	logger := k.Logger(ctx)

	// Open commitment store
	store := k.storeService.OpenKVStore(ctx)
	commitmentStore := prefix.NewStore(runtime.KVStoreAdapter(store), types.CommitmentKey)

	// Iterate and delete commitments
	iterator := commitmentStore.Iterator(nil, nil)
	defer func() {
		if err := iterator.Close(); err != nil {
			logger.Error("failed to close commitment iterator",
				"period", period,
				"error", err)
		}
	}()

	for ; iterator.Valid(); iterator.Next() {
		if err := store.Delete(iterator.Key()); err != nil {
			logger.Error("failed to delete commitment",
				"period", period,
				"key", iterator.Key(),
				"error", err)
			return sdkerrors.Wrap(errortypes.ErrInvalidState, "failed to delete commitment")
		}
	}

	// Open reveal store
	store = k.storeService.OpenKVStore(ctx)
	revealStore := prefix.NewStore(runtime.KVStoreAdapter(store), types.RevealKey)

	// Iterate and delete reveals
	iterator = revealStore.Iterator(nil, nil)
	defer func() {
		if err := iterator.Close(); err != nil {
			logger.Error("failed to close reveal iterator",
				"period", period,
				"error", err)
		}
	}()

	for ; iterator.Valid(); iterator.Next() {
		if err := store.Delete(iterator.Key()); err != nil {
			logger.Error("failed to delete reveal",
				"period", period,
				"key", iterator.Key(),
				"error", err)
			return sdkerrors.Wrap(errortypes.ErrInvalidState, "failed to delete reveal")
		}
	}

	logger.Info("period data cleared successfully",
		"period", period)

	return nil
}

func (k Keeper) Commit(goCtx context.Context, period uint64, validatorAddress string, deposit sdk.Coin, commitHash []byte, delegationAddress string) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	logger := ctx.Logger().With("module", "keeper")

	// Get current period state
	state, found := k.GetPeriod(ctx)
	if !found {
		logger.Error("failed to get period state",
			"found", found,
			"validator", validatorAddress,
			"period", period)
		return sdkerrors.Wrap(errortypes.ErrInvalidState, "failed to get period state")
	}

	// Verify commit phase
	if !state.InCommitPhase {
		logger.Error("invalid phase for commitment",
			"validator", validatorAddress,
			"period", period)
		return sdkerrors.Wrap(errortypes.ErrInvalidPhase, "not in commit phase")
	}

	if ctx.BlockHeight() >= state.CommitEndHeight {
		logger.Error("commit phase closed",
			"validator", validatorAddress,
			"period", period,
			"block_height", ctx.BlockHeight())
		return sdkerrors.Wrap(errortypes.ErrCommitPhaseClosed, "commit phase closed")
	}

	// Parse validator address
	valAddress, err := sdk.ValAddressFromBech32(validatorAddress)
	if err != nil {
		logger.Error("invalid validator address format",
			"validator", validatorAddress,
			"error", err)
		return sdkerrors.Wrap(errortypes.ErrInvalidValidator, "invalid validator address format")
	}

	// Verify validator
	validator, err := k.stakingKeeper.Validator(ctx, valAddress)
	if err != nil || validator == nil || !validator.IsBonded() {
		logger.Error("invalid or inactive validator",
			"validator", validatorAddress,
			"error", err)
		return sdkerrors.Wrap(errortypes.ErrInvalidValidator, "invalid or inactive validator")
	}

	// Verify no duplicate commitment
	if k.HasCommitment(ctx, state.CurrentPeriod, validatorAddress) {
		logger.Error("duplicate commitment detected",
			"validator", validatorAddress,
			"period", state.CurrentPeriod)
		return sdkerrors.Wrap(errortypes.ErrDuplicateCommitment, "validator already committed")
	}

	// Verify deposit
	params := k.GetParams(ctx)
	if !deposit.IsGTE(*params.MinimumDeposit) {
		logger.Error("insufficient deposit",
			"validator", validatorAddress,
			"period", state.CurrentPeriod,
			"deposit", deposit.String())
		return sdkerrors.Wrap(errortypes.ErrInsufficientDeposit, "insufficient deposit")
	}

	// Prepare and transfer tokens
	token := sdk.NewCoins()
	token = token.Add(sdk.NewCoin(appParams.BondDenom, math.NewInt(deposit.Amount.Int64())))

	senderAddress, err := sdk.AccAddressFromBech32(delegationAddress)
	if err != nil {
		logger.Error("failed to parse sender address",
			"address", delegationAddress,
			"error", err)
		return sdkerrors.Wrap(errortypes.ErrInvalidAddress, "failed to parse sender address")
	}

	logger.Info("successfully parsed sender address",
		"address", delegationAddress)

	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, sdk.AccAddress(senderAddress), types.ModuleName, sdk.NewCoins(token...))
	if err != nil {
		logger.Error("failed to transfer deposit",
			"validator", validatorAddress,
			"period", state.CurrentPeriod,
			"error", err)
		return sdkerrors.Wrapf(errortypes.ErrInvalidCoins, "deposit did not transfer: %s", err)
	}

	// Store commitment
	commitment := types.Commitment{
		Period:         period,
		Validator:      validatorAddress,
		CommitmentHash: commitHash,
		BlockHeight:    ctx.BlockHeight(),
		Revealed:       false,
		Deposit:        &deposit,
	}
	k.SetCommitment(ctx, commitment)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCommitment,
			sdk.NewAttribute(types.AttributeKeyValidator, validatorAddress),
			sdk.NewAttribute(types.AttributeKeyPeriodNumber, fmt.Sprintf("%d", state.CurrentPeriod)),
		),
	)

	logger.Info("commitment successful",
		"validator", validatorAddress,
		"period", state.CurrentPeriod,
		"block_height", ctx.BlockHeight())

	return nil
}

func (k Keeper) Reveal(goCtx context.Context, validator string, revealPeriod uint64, delegationAddress string) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	logger := ctx.Logger().With("module", "keeper")

	// Get current period state
	state, found := k.GetPeriod(ctx)
	if !found {
		logger.Error("failed to get period state",
			"validator", validator,
			"period", revealPeriod,
			"found", found)
		return sdkerrors.Wrap(errortypes.ErrInvalidState, "failed to get period state")
	}

	// Verify reveal phase
	if state.InCommitPhase {
		logger.Error("invalid phase for reveal",
			"validator", validator,
			"period", revealPeriod)
		return sdkerrors.Wrap(errortypes.ErrInvalidPhase, "not in reveal phase")
	}

	if ctx.BlockHeight() >= state.RevealEndHeight {
		logger.Error("reveal phase closed",
			"validator", validator,
			"period", revealPeriod,
			"block_height", ctx.BlockHeight())
		return sdkerrors.Wrap(errortypes.ErrRevealPhaseClosed, "reveal phase closed")
	}

	// Get commitment
	commitment, found := k.GetCommitment(ctx, revealPeriod, validator)
	if !found {
		logger.Error("no commitment found",
			"validator", validator,
			"period", revealPeriod)
		return sdkerrors.Wrap(errortypes.ErrNoCommitment, "no commitment found")
	}

	if !commitment.Revealed {
		// Generate and verify random value
		r := k.GenerateRandomValue(sdk.ValAddress(validator), revealPeriod)
		computedHash := k.ComputeHash(r)

		if !bytes.Equal(computedHash, commitment.CommitmentHash) {
			logger.Error("reveal does not match commitment",
				"validator", validator,
				"period", revealPeriod)
			return sdkerrors.Wrap(errortypes.ErrInvalidReveal, "reveal does not match commitment")
		}

		// Mark commitment as revealed
		commitment.Revealed = true
		k.SetCommitment(ctx, commitment)

		// Convert random value to bytes
		revealValue, err := k.ConvertRandomValueToBytes(r)
		if err != nil {
			logger.Error("failed to convert random value to bytes",
				"validator", validator,
				"period", revealPeriod,
				"error", err)
			return sdkerrors.Wrap(errortypes.ErrInvalidReveal, "failed to convert random value")
		}

		// Store reveal
		reveal := types.Reveal{
			Period:      state.CurrentPeriod,
			Validator:   validator,
			RevealValue: revealValue,
			BlockHeight: ctx.BlockHeight(),
		}
		k.SetReveal(ctx, reveal)

		// Prepare tokens for deposit return
		token := sdk.NewCoins()
		token = token.Add(sdk.NewCoin(appParams.BondDenom, math.NewInt(commitment.Deposit.Amount.Int64())))

		senderAddress, err := sdk.AccAddressFromBech32(delegationAddress)
		if err != nil {
			logger.Error("failed to parse sender address",
				"address", delegationAddress,
				"error", err)
			return sdkerrors.Wrap(errortypes.ErrInvalidAddress, "failed to parse sender address")
		}

		logger.Info("successfully parsed sender address",
			"address", delegationAddress)

		// Return deposit
		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, senderAddress, sdk.NewCoins(token...))
		if err != nil {
			logger.Error("failed to return deposit",
				"validator", validator,
				"period", revealPeriod,
				"error", err)
			return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "deposit can not transfer from module to account")
		}

		// Emit event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeReveal,
				sdk.NewAttribute(types.AttributeKeyValidator, validator),
				sdk.NewAttribute(types.AttributeKeyPeriodNumber, fmt.Sprintf("%d", state.CurrentPeriod)),
			),
		)
	}

	logger.Info("reveal successful",
		"validator", validator,
		"period", state.CurrentPeriod,
		"block_height", ctx.BlockHeight())

	return nil
}
