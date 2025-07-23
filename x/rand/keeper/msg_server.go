package keeper

import (
	"bytes"
	"context"
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	appParams "github.com/outbe/outbe-node/app/params"
	errortypes "github.com/outbe/outbe-node/errors"
	"github.com/outbe/outbe-node/x/rand/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// Commit handles MsgCommit
func (k msgServer) Commit(goCtx context.Context, msg *types.MsgCommit) (*types.MsgCommitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	state, _ := k.GetPeriod(ctx)

	logger := k.Logger(ctx)

	// Verify commit phase
	if !state.InCommitPhase {
		return nil, sdkerrors.Wrap(errortypes.ErrInvalidPhase, "not in commit phase")
	}
	if ctx.BlockHeight() >= state.CommitEndHeight {
		return nil, sdkerrors.Wrap(errortypes.ErrCommitPhaseClosed, "commit phase closed")
	}

	if len(msg.CommitmentHash) != 32 { // Assuming SHA-256
		return nil, sdkerrors.Wrap(errortypes.ErrInvalidHash, "commitment hash must be 32 bytes")
	}

	valAddress, _ := sdk.ValAddressFromBech32(msg.Validator)

	// Verify validator
	validator, err := k.stakingKeeper.Validator(ctx, valAddress)
	if err != nil || validator == nil || validator.IsJailed() || validator.IsUnbonded() || validator.IsUnbonding() {
		logger.Error("invalid or inactive validator",
			"validator", msg.Validator,
			"error", err)
		return nil, sdkerrors.Wrap(errortypes.ErrInvalidValidator, "invalid or inactive validator")
	}

	// Verify no duplicate commitment
	if k.HasCommitment(ctx, state.CurrentPeriod, msg.Validator) {
		return nil, sdkerrors.Wrap(errortypes.ErrDuplicateCommitment, "validator already committed")
	}

	// Verify deposit
	params := k.GetParams(ctx)
	if !msg.Deposit.IsGTE(*params.MinimumDeposit) {
		return nil, sdkerrors.Wrap(errortypes.ErrInsufficientDeposit, "insufficient deposit")
	}

	token := sdk.NewCoins()
	token = token.Add(sdk.NewCoin(appParams.BondDenom, math.NewInt(msg.Deposit.Amount.Int64())))

	senderAddress, _ := sdk.AccAddressFromBech32(msg.Creator)

	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddress, types.ModuleName, sdk.NewCoins(token...))
	if err != nil {
		return nil, sdkerrors.Wrap(errortypes.ErrInvalidCoins, "deposit did not transfered")
	}

	// Store commitment
	commitment := types.Commitment{
		Period:         state.CurrentPeriod,
		Validator:      msg.Validator,
		CommitmentHash: msg.CommitmentHash,
		BlockHeight:    ctx.BlockHeight(),
		Revealed:       false,
		Deposit:        msg.Deposit,
	}
	k.SetCommitment(ctx, commitment)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCommitment,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.Validator),
			sdk.NewAttribute(types.AttributeKeyPeriodNumber, fmt.Sprintf("%d", state.CurrentPeriod)),
		),
	)

	return &types.MsgCommitResponse{}, nil
}

// Reveal handles MsgReveal
func (k msgServer) Reveal(goCtx context.Context, msg *types.MsgReveal) (*types.MsgRevealResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	state, _ := k.GetPeriod(ctx)

	logger := k.Logger(ctx)

	// Verify reveal phase
	if state.InCommitPhase {
		return nil, sdkerrors.Wrap(errortypes.ErrInvalidPhase, "not in reveal phase")
	}

	if ctx.BlockHeight() >= state.RevealEndHeight {
		return nil, sdkerrors.Wrap(errortypes.ErrRevealPhaseClosed, "reveal phase closed")
	}

	_, found := k.GetPenalty(ctx, msg.Period, msg.Validator)
	if found {
		return nil, sdkerrors.Wrap(errortypes.ErrHasPenalty, "find validator penalty")
	}

	commitment, found := k.GetCommitment(ctx, msg.Period, msg.Validator)
	if !found {
		return nil, sdkerrors.Wrap(errortypes.ErrNoCommitment, "no commitment found")
	}

	if commitment.Revealed {
		return nil, sdkerrors.Wrap(errortypes.ErrAlreadyRevealed, "already revealed")
	}

	computedHash := ComputeHash(msg.RevealValue)
	if !bytes.Equal(computedHash, commitment.CommitmentHash) {
		logger.Error("reveal does not match commitment",
			"validator", msg.Validator,
			"period", msg.Period,
			"reveal_value", fmt.Sprintf("%x", msg.RevealValue),
			"computed_hash", fmt.Sprintf("%x", computedHash),
			"commitment_hash", fmt.Sprintf("%x", commitment.CommitmentHash))
		return nil, sdkerrors.Wrap(errortypes.ErrInvalidReveal, "reveal does not match commitment")
	}

	// Mark commitment as revealed
	commitment.Revealed = true
	k.SetCommitment(ctx, commitment)

	// Store reveal
	reveal := types.Reveal{
		Period:      state.CurrentPeriod,
		Validator:   msg.Validator,
		RevealValue: msg.RevealValue,
		BlockHeight: ctx.BlockHeight(),
	}
	k.SetReveal(ctx, reveal)

	token := sdk.NewCoins()
	token = token.Add(sdk.NewCoin(appParams.BondDenom, math.NewInt(commitment.Deposit.Amount.Int64())))

	senderAddress, _ := sdk.AccAddressFromBech32(msg.Creator)

	// Return deposit
	err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, senderAddress, sdk.NewCoins(token...))
	if err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeReveal,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.Validator),
			sdk.NewAttribute(types.AttributeKeyPeriodNumber, fmt.Sprintf("%d", state.CurrentPeriod)),
		),
	)

	return &types.MsgRevealResponse{}, nil
}
