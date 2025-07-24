package keeper

import (
	"context"
	"strconv"
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/gogo/status"
	"google.golang.org/grpc/codes"
)

type msgServer struct {
	WrappedBaseKeeper
}

func NewMsgServerImpl(keeper WrappedBaseKeeper) types.MsgServer {
	return &msgServer{WrappedBaseKeeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) CreateValidator(ctx context.Context, msg *types.MsgCreateValidator) (*types.MsgCreateValidatorResponse, error) {
	goctx := sdk.UnwrapSDKContext(ctx)
	params := k.rk.GetParams(goctx)

	maxSelfBondToken, _ := math.NewIntFromString(params.MaxSelfBondToken)

	validators, err := k.GetAllValidators(ctx)
	if err != nil {
		return nil, sdkerrors.ErrInvalidType.Wrapf("failed to query validators")
	}

	if len(validators) > 0 {
		if msg.Value.Amount.GT(maxSelfBondToken) {
			return nil, sdkerrors.ErrInvalidCoins.Wrapf("validator self bond token exceeded maximum. it should be lower than: %s", params.MaxSelfBondToken)
		}
	}

	valAddr, err := k.Keeper.ValidatorAddressCodec().StringToBytes(msg.ValidatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	if err := msg.Validate(k.Keeper.ValidatorAddressCodec()); err != nil {
		return nil, err
	}

	minCommRate, err := k.MinCommissionRate(ctx)
	if err != nil {
		return nil, err
	}

	if msg.Commission.Rate.LT(minCommRate) {
		return nil, errorsmod.Wrapf(types.ErrCommissionLTMinRate, "cannot set validator commission to less than minimum rate of %s", minCommRate)
	}

	// check to see if the pubkey or sender has been registered before
	if _, err := k.GetValidator(ctx, valAddr); err == nil {
		return nil, types.ErrValidatorOwnerExists
	}

	pk, ok := msg.Pubkey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidType, "Expecting cryptotypes.PubKey, got %T", pk)
	}

	if _, err := k.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(pk)); err == nil {
		return nil, types.ErrValidatorPubKeyExists
	}

	bondDenom, err := k.BondDenom(ctx)
	if err != nil {
		return nil, err
	}

	if msg.Value.Denom != bondDenom {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Value.Denom, bondDenom,
		)
	}

	if _, err := msg.Description.EnsureLength(); err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	cp := sdkCtx.ConsensusParams()
	if cp.Validator != nil {
		pkType := pk.Type()
		hasKeyType := false
		for _, keyType := range cp.Validator.PubKeyTypes {
			if pkType == keyType {
				hasKeyType = true
				break
			}
		}
		if !hasKeyType {
			return nil, errorsmod.Wrapf(
				types.ErrValidatorPubKeyTypeNotSupported,
				"got: %s, expected: %s", pk.Type(), cp.Validator.PubKeyTypes,
			)
		}
	}

	validator, err := types.NewValidator(msg.ValidatorAddress, pk, msg.Description)
	if err != nil {
		return nil, err
	}

	commission := types.NewCommissionWithTime(
		msg.Commission.Rate, msg.Commission.MaxRate,
		msg.Commission.MaxChangeRate, sdkCtx.BlockHeader().Time,
	)

	validator, err = validator.SetInitialCommission(commission)
	if err != nil {
		return nil, err
	}

	validator.MinSelfDelegation = msg.MinSelfDelegation

	err = k.SetValidator(ctx, validator)
	if err != nil {
		return nil, err
	}

	err = k.SetValidatorByConsAddr(ctx, validator)
	if err != nil {
		return nil, err
	}

	err = k.SetNewValidatorByPowerIndex(ctx, validator)
	if err != nil {
		return nil, err
	}

	// call the after-creation hook
	if err := k.Hooks().AfterValidatorCreated(ctx, valAddr); err != nil {
		return nil, err
	}

	// move coins from the msg.Address account to a (self-delegation) delegator account
	// the validator account and global shares are updated within here
	// NOTE source will always be from a wallet which are unbonded
	_, err = k.Keeper.Delegate(ctx, sdk.AccAddress(valAddr), msg.Value.Amount, types.Unbonded, validator, true)
	if err != nil {
		return nil, err
	}

	sdkCtx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreateValidator,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Value.String()),
		),
	})

	return &types.MsgCreateValidatorResponse{}, nil
}

func (k msgServer) EditValidator(ctx context.Context, msg *types.MsgEditValidator) (*types.MsgEditValidatorResponse, error) {
	goctx := sdk.UnwrapSDKContext(ctx)
	params := k.rk.GetParams(goctx)

	valAddr, err := k.Keeper.ValidatorAddressCodec().StringToBytes(msg.ValidatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	maxSelfBond, err := strconv.ParseInt(params.MaxSelfBondToken, 10, 64)
	if err == nil {
		return nil, sdkerrors.ErrInvalidType.Wrapf("failed to convert MaxSelfBondToken amount from string to int")
	}

	maxSelfBondToken := math.NewInt(maxSelfBond)

	validators, err := k.GetAllValidators(ctx)
	if err != nil {
		return nil, sdkerrors.ErrInvalidType.Wrapf("failed to query validators")
	}
	if len(validators) > 0 {
		if msg.MinSelfDelegation.GT(maxSelfBondToken) {
			return nil, sdkerrors.ErrInvalidCoins.Wrapf("validator self bond token exceeded maximum. it should be lower than: %s", params.MaxSelfBondToken)
		}
	}

	if msg.Description == (types.Description{}) {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty description")
	}

	if msg.MinSelfDelegation != nil && !msg.MinSelfDelegation.IsPositive() {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"minimum self delegation must be a positive integer",
		)
	}

	if msg.CommissionRate != nil {
		if msg.CommissionRate.GT(math.LegacyOneDec()) || msg.CommissionRate.IsNegative() {
			return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "commission rate must be between 0 and 1 (inclusive)")
		}

		minCommissionRate, err := k.MinCommissionRate(ctx)
		if err != nil {
			return nil, errorsmod.Wrap(sdkerrors.ErrLogic, err.Error())
		}

		if msg.CommissionRate.LT(minCommissionRate) {
			return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "commission rate cannot be less than the min commission rate %s", minCommissionRate.String())
		}
	}

	// validator must already be registered
	validator, err := k.GetValidator(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	// replace all editable fields (clients should autofill existing values)
	description, err := validator.Description.UpdateDescription(msg.Description)
	if err != nil {
		return nil, err
	}

	validator.Description = description

	if msg.CommissionRate != nil {
		commission, err := k.UpdateValidatorCommission(ctx, validator, *msg.CommissionRate)
		if err != nil {
			return nil, err
		}

		// call the before-modification hook since we're about to update the commission
		if err := k.Hooks().BeforeValidatorModified(ctx, valAddr); err != nil {
			return nil, err
		}

		validator.Commission = commission
	}

	if msg.MinSelfDelegation != nil {
		if !msg.MinSelfDelegation.GT(validator.MinSelfDelegation) {
			return nil, types.ErrMinSelfDelegationDecreased
		}

		if msg.MinSelfDelegation.GT(validator.Tokens) {
			return nil, types.ErrSelfDelegationBelowMinimum
		}

		validator.MinSelfDelegation = *msg.MinSelfDelegation
	}

	err = k.SetValidator(ctx, validator)
	if err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeEditValidator,
			sdk.NewAttribute(types.AttributeKeyCommissionRate, validator.Commission.String()),
			sdk.NewAttribute(types.AttributeKeyMinSelfDelegation, validator.MinSelfDelegation.String()),
		),
	})

	return &types.MsgEditValidatorResponse{}, nil
}

func (k msgServer) Delegate(ctx context.Context, msg *types.MsgDelegate) (*types.MsgDelegateResponse, error) {
	valAddr, valErr := k.Keeper.ValidatorAddressCodec().StringToBytes(msg.ValidatorAddress)
	if valErr != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", valErr)
	}

	delegatorAddress, err := k.ak.AddressCodec().StringToBytes(msg.DelegatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid delegator address: %s", err)
	}

	if !msg.Amount.IsValid() || !msg.Amount.Amount.IsPositive() {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"invalid delegation amount",
		)
	}

	validator, err := k.GetValidator(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	bondDenom, err := k.BondDenom(ctx)
	if err != nil {
		return nil, err
	}

	if msg.Amount.Denom != bondDenom {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Amount.Denom, bondDenom,
		)
	}

	// NOTE: source funds are always unbonded
	newShares, err := k.Keeper.Delegate(ctx, delegatorAddress, msg.Amount.Amount, types.Unbonded, validator, true)
	if err != nil {
		return nil, err
	}

	if msg.Amount.Amount.IsInt64() {
		defer func() {
			telemetry.IncrCounter(1, types.ModuleName, "delegate")
		}()
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeDelegate,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
			sdk.NewAttribute(types.AttributeKeyDelegator, msg.DelegatorAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyNewShares, newShares.String()),
		),
	})

	return &types.MsgDelegateResponse{}, nil
}

func (k msgServer) BeginRedelegate(ctx context.Context, msg *types.MsgBeginRedelegate) (*types.MsgBeginRedelegateResponse, error) {
	valSrcAddr, err := k.Keeper.ValidatorAddressCodec().StringToBytes(msg.ValidatorSrcAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid source validator address: %s", err)
	}

	valDstAddr, err := k.Keeper.ValidatorAddressCodec().StringToBytes(msg.ValidatorDstAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid destination validator address: %s", err)
	}

	delegatorAddress, err := k.ak.AddressCodec().StringToBytes(msg.DelegatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid delegator address: %s", err)
	}

	if !msg.Amount.IsValid() || !msg.Amount.Amount.IsPositive() {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"invalid shares amount",
		)
	}

	shares, err := k.ValidateUnbondAmount(
		ctx, delegatorAddress, valSrcAddr, msg.Amount.Amount,
	)
	if err != nil {
		return nil, err
	}

	bondDenom, err := k.BondDenom(ctx)
	if err != nil {
		return nil, err
	}

	if msg.Amount.Denom != bondDenom {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Amount.Denom, bondDenom,
		)
	}

	completionTime, err := k.BeginRedelegation(
		ctx, delegatorAddress, valSrcAddr, valDstAddr, shares,
	)
	if err != nil {
		return nil, err
	}

	if msg.Amount.Amount.IsInt64() {
		defer func() {
			telemetry.IncrCounter(1, types.ModuleName, "redelegate")
		}()
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeRedelegate,
			sdk.NewAttribute(types.AttributeKeySrcValidator, msg.ValidatorSrcAddress),
			sdk.NewAttribute(types.AttributeKeyDstValidator, msg.ValidatorDstAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyCompletionTime, completionTime.Format(time.RFC3339)),
		),
	})

	return &types.MsgBeginRedelegateResponse{
		CompletionTime: completionTime,
	}, nil
}

func (k msgServer) Undelegate(ctx context.Context, msg *types.MsgUndelegate) (*types.MsgUndelegateResponse, error) {
	addr, err := k.Keeper.ValidatorAddressCodec().StringToBytes(msg.ValidatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	delegatorAddress, err := k.ak.AddressCodec().StringToBytes(msg.DelegatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid delegator address: %s", err)
	}

	if !msg.Amount.IsValid() || !msg.Amount.IsPositive() {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"invalid shares amount",
		)
	}

	shares, err := k.Keeper.ValidateUnbondAmount(
		ctx, delegatorAddress, addr, msg.Amount.Amount,
	)
	if err != nil {
		return nil, err
	}

	bondDenom, err := k.Keeper.BondDenom(ctx)
	if err != nil {
		return nil, err
	}

	if msg.Amount.Denom != bondDenom {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Amount.Denom, bondDenom,
		)
	}

	completionTime, undelegatedAmt, err := k.Keeper.Undelegate(ctx, delegatorAddress, addr, shares)
	if err != nil {
		return nil, err
	}

	undelegatedCoin := sdk.NewCoin(msg.Amount.Denom, undelegatedAmt)

	if msg.Amount.Amount.IsInt64() {
		defer func() {
			telemetry.IncrCounter(1, types.ModuleName, "undelegate")
		}()
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeUnbond,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
			sdk.NewAttribute(types.AttributeKeyDelegator, msg.DelegatorAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, undelegatedCoin.String()),
			sdk.NewAttribute(types.AttributeKeyCompletionTime, completionTime.Format(time.RFC3339)),
		),
	})

	return &types.MsgUndelegateResponse{
		CompletionTime: completionTime,
		Amount:         undelegatedCoin,
	}, nil
}

func (k msgServer) CancelUnbondingDelegation(ctx context.Context, msg *types.MsgCancelUnbondingDelegation) (*types.MsgCancelUnbondingDelegationResponse, error) {
	valAddr, err := k.Keeper.ValidatorAddressCodec().StringToBytes(msg.ValidatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	delegatorAddress, err := k.ak.AddressCodec().StringToBytes(msg.DelegatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid delegator address: %s", err)
	}

	if !msg.Amount.IsValid() || !msg.Amount.Amount.IsPositive() {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"invalid amount",
		)
	}

	if msg.CreationHeight <= 0 {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"invalid height",
		)
	}

	bondDenom, err := k.BondDenom(ctx)
	if err != nil {
		return nil, err
	}

	if msg.Amount.Denom != bondDenom {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Amount.Denom, bondDenom,
		)
	}

	validator, err := k.GetValidator(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	// In some situations, the exchange rate becomes invalid, e.g. if
	// Validator loses all tokens due to slashing. In this case,
	// make all future delegations invalid.
	if validator.InvalidExRate() {
		return nil, types.ErrDelegatorShareExRateInvalid
	}

	if validator.IsJailed() {
		return nil, types.ErrValidatorJailed
	}

	ubd, err := k.GetUnbondingDelegation(ctx, delegatorAddress, valAddr)
	if err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			"unbonding delegation with delegator %s not found for validator %s",
			msg.DelegatorAddress, msg.ValidatorAddress,
		)
	}

	var (
		unbondEntry      types.UnbondingDelegationEntry
		unbondEntryIndex int64 = -1
	)

	for i, entry := range ubd.Entries {
		if entry.CreationHeight == msg.CreationHeight {
			unbondEntry = entry
			unbondEntryIndex = int64(i)
			break
		}
	}
	if unbondEntryIndex == -1 {
		return nil, sdkerrors.ErrNotFound.Wrapf("unbonding delegation entry is not found at block height %d", msg.CreationHeight)
	}

	if unbondEntry.Balance.LT(msg.Amount.Amount) {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("amount is greater than the unbonding delegation entry balance")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if unbondEntry.CompletionTime.Before(sdkCtx.BlockTime()) {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("unbonding delegation is already processed")
	}

	// delegate back the unbonding delegation amount to the validator
	_, err = k.Keeper.Delegate(ctx, delegatorAddress, msg.Amount.Amount, types.Unbonding, validator, false)
	if err != nil {
		return nil, err
	}

	amount := unbondEntry.Balance.Sub(msg.Amount.Amount)
	if amount.IsZero() {
		ubd.RemoveEntry(unbondEntryIndex)
	} else {
		// update the unbondingDelegationEntryBalance and InitialBalance for ubd entry
		unbondEntry.Balance = amount
		unbondEntry.InitialBalance = unbondEntry.InitialBalance.Sub(msg.Amount.Amount)
		ubd.Entries[unbondEntryIndex] = unbondEntry
	}

	// set the unbonding delegation or remove it if there are no more entries
	if len(ubd.Entries) == 0 {
		err = k.RemoveUnbondingDelegation(ctx, ubd)
	} else {
		err = k.SetUnbondingDelegation(ctx, ubd)
	}

	if err != nil {
		return nil, err
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCancelUnbondingDelegation,
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
			sdk.NewAttribute(types.AttributeKeyDelegator, msg.DelegatorAddress),
			sdk.NewAttribute(types.AttributeKeyCreationHeight, strconv.FormatInt(msg.CreationHeight, 10)),
		),
	)

	return &types.MsgCancelUnbondingDelegationResponse{}, nil
}

func (k msgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if k.Keeper.GetAuthority() != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.Keeper.GetAuthority(), msg.Authority)
	}

	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}

	// store params
	if err := k.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

func (k msgServer) getValidator(goCtx context.Context, address string) (types.Validator, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	logger := k.Logger(ctx)
	validators, err := k.Keeper.GetAllValidators(ctx)
	if err != nil {
		logger.Error("[Undelegate] Minting failed",
			"validator address", address,
			"error", err,
		)
		return types.Validator{}, err
	}
	for _, validator := range validators {
		if validator.OperatorAddress == address {
			return validator, nil
		}
	}
	return types.Validator{}, nil
}
