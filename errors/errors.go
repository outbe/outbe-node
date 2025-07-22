package types

import (
	sdkerrors "cosmossdk.io/errors"
	"google.golang.org/grpc/codes"
)

var ModuleName string

var (
	ConvertToString = "[Type Error] Conversion failed: cannot convert from Int to String."
	ParseUint       = "[Parse Error] Failed to parse the provided data. Please ensure the input format is correct."
	ErrEmptyAddress = "[Address Error] The provided address is empty. AccAddressFromBech32 requires a non-empty address string."
	InvalidRequest  = "[Request Error] The request is invalid. Please check the parameters and try again."
)

var (
	ErrInvalidRequest = sdkerrors.Register(ModuleName, 1, "Invalid Request: The request format is invalid or missing required parameters.")
	ErrNotFound       = sdkerrors.RegisterWithGRPCCode(ModuleName, 2, codes.NotFound, "not found")
)

var (
	ErrKeyNotFound = sdkerrors.Register(ModuleName, 3, "Key Not Found: The specified key could not be located in the data store.")
)

var (
	ErrInvalidType    = sdkerrors.Register(ModuleName, 4, "Invalid Type: The provided value is of an unexpected type and cannot be processed.")
	ErrUnknownRequest = sdkerrors.Register(ModuleName, 5, "Unknown Request: The request type is unrecognized. Please verify the request details.")
)

var (
	ErrInvalidMintAmount = sdkerrors.Register(ModuleName, 6, "Invalid Mint Amount: The mint amount must be greater than zero.")
	ErrJSONUnmarshal     = sdkerrors.Register(ModuleName, 7, "failed to unmarshal JSON bytes")
)

var (
	ErrCalculation           = sdkerrors.Register(ModuleName, 8, "failed to calculate ")
	ErrUnauthorized          = sdkerrors.Register(ModuleName, 9, "unauthorized operation")
	ErrConflict              = sdkerrors.Register(ModuleName, 10, "conflict: entry already exists")
	ErrNoValidatorForAddress = sdkerrors.Register(ModuleName, 11, "address is not associated with any known validator")
	ErrInsufficientFunds     = sdkerrors.Register(ModuleName, 12, "no insufficient funds")
	ErrInvalidCoins          = sdkerrors.Register(ModuleName, 13, "no valid coin")
)

var (
	ErrInvalidPhase        = sdkerrors.Register(ModuleName, 14, "failed to get valid phase ")
	ErrCommitPhaseClosed   = sdkerrors.Register(ModuleName, 15, "closing commit phase")
	ErrInvalidValidator    = sdkerrors.Register(ModuleName, 16, "invalid validator")
	ErrDuplicateCommitment = sdkerrors.Register(ModuleName, 17, "duplicate commitment")
	ErrInsufficientDeposit = sdkerrors.Register(ModuleName, 18, "insufficient deposit")
	ErrAlreadyRevealed     = sdkerrors.Register(ModuleName, 19, "already revealed")
	ErrNoCommitment        = sdkerrors.Register(ModuleName, 20, "no commitment")
	ErrRevealPhaseClosed   = sdkerrors.Register(ModuleName, 21, "closing reveal phase")
	ErrInvalidReveal       = sdkerrors.Register(ModuleName, 22, "invalid reveal")
)

var (
	ErrInvalidAddress = sdkerrors.Register(ModuleName, 23, "invalid address")
	ErrInvalidState   = sdkerrors.Register(ModuleName, 24, "failed to get period stat")
)
