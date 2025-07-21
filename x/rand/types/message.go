package types

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/outbe/outbe-node/errors"
)

var (
	_ sdk.Msg = &MsgCommit{}
	_ sdk.Msg = &MsgReveal{}
)

// NewMsgCommit creates a new MsgCommit
func NewMsgCommit(creator string, validator sdk.ValAddress, commitmentHash string, deposit sdk.Coin) *MsgCommit {
	return &MsgCommit{
		Creator:        creator,
		Validator:      validator.String(),
		CommitmentHash: commitmentHash,
		Deposit:        &deposit,
	}
}

// Route implements sdk.Msg
func (msg MsgCommit) Route() string { return RouterKey }

// Type implements sdk.Msg
func (msg MsgCommit) Type() string { return "commit" }

// ValidateBasic implements sdk.Msg
func (msg MsgCommit) ValidateBasic() error {
	if msg.Validator == "" {
		return sdkerrors.Wrap(errortypes.ErrInvalidAddress, "validator address cannot be empty")
	}
	if len(msg.CommitmentHash) == 0 {
		return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "commitment hash cannot be empty")
	}
	if !msg.Deposit.IsValid() {
		return sdkerrors.Wrap(errortypes.ErrInvalidCoins, "invalid deposit")
	}
	return nil
}

// GetSigners implements sdk.Msg
func (msg MsgCommit) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.AccAddress(msg.Validator)}
}

// NewMsgReveal creates a new MsgReveal
func NewMsgReveal(creator string, period string, validator string, revealValue string) *MsgReveal {
	return &MsgReveal{
		Creator:     creator,
		Validator:   validator,
		RevealValue: revealValue,
		Period:      period,
	}
}

// Route implements sdk.Msg
func (msg MsgReveal) Route() string { return RouterKey }

// Type implements sdk.Msg
func (msg MsgReveal) Type() string { return "reveal" }

// ValidateBasic implements sdk.Msg
func (msg MsgReveal) ValidateBasic() error {
	if msg.Validator == "" {
		return sdkerrors.Wrap(errortypes.ErrInvalidAddress, "validator address cannot be empty")
	}
	if len(msg.RevealValue) == 0 {
		return sdkerrors.Wrap(errortypes.ErrInvalidRequest, "reveal value cannot be empty")
	}
	return nil
}

// GetSigners implements sdk.Msg
func (msg MsgReveal) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.AccAddress(msg.Validator)}
}
