package wasmbinding

import (
	"encoding/json"
	"math"
	"strconv"

	errorsmod "cosmossdk.io/errors"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/outbe/outbe-node/wasmbinding/bindings"

	poolkeeper "github.com/outbe/outbe-node/x/pool/keeper"
	pooltypes "github.com/outbe/outbe-node/x/pool/types"

	errortypes "github.com/outbe/outbe-node/errors"
)

// CustomMessageDecorator returns decorator for custom CosmWasm bindings messages
func CustomMessageDecorator(bank *bankkeeper.BaseKeeper, poolMint *poolkeeper.Keeper) func(wasmkeeper.Messenger) wasmkeeper.Messenger {
	return func(old wasmkeeper.Messenger) wasmkeeper.Messenger {
		return &CustomMessenger{
			wrapped:  old,
			bank:     bank,
			poolMint: poolMint,
		}
	}
}

type CustomMessenger struct {
	wrapped  wasmkeeper.Messenger
	bank     *bankkeeper.BaseKeeper
	poolMint *poolkeeper.Keeper
}

var _ wasmkeeper.Messenger = (*CustomMessenger)(nil)

// DispatchMsg executes on the contractMsg.
func (m *CustomMessenger) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) (events []sdk.Event, data [][]byte, msgResponses [][]*codectypes.Any, err error) {

	// only handle the happy path where this is really minting ...
	// leave everything else for the wrapped version
	var redumptionMsg bindings.QcoreMsg
	if err := json.Unmarshal(msg.Custom, &redumptionMsg); err != nil {
		return nil, nil, nil, errorsmod.Wrap(err, "[DispatchMsg][Unmarshal failed to unmarshal the message")
	}
	if redumptionMsg.MsgMintTribute != nil {
		return m.mintTokens(ctx, contractAddr, redumptionMsg.MsgMintTribute)
	}
	return m.wrapped.DispatchMsg(ctx, contractAddr, contractIBCPortID, msg)
}

// mintTokens mints tokens of a specified denom to an address.
func (m *CustomMessenger) mintTokens(ctx sdk.Context, contractAddr sdk.AccAddress, mint *bindings.MsgMintTribute) (events []sdk.Event, data [][]byte, msgResponses [][]*codectypes.Any, err error) {

	err = PerformMint(m.poolMint, m.bank, ctx, contractAddr, mint)

	if err != nil {
		return nil, nil, nil, errorsmod.Wrap(err, "[mintTokens] failed to perform mint")
	}
	return nil, nil, nil, nil
}

// PerformMint used with mintTokens to validate the mint message and mint through token factory.
func PerformMint(f *poolkeeper.Keeper, b *bankkeeper.BaseKeeper, ctx sdk.Context, contractAddr sdk.AccAddress, bindingMsg *bindings.MsgMintTribute) error {

	if bindingMsg == nil {
		return wasmvmtypes.InvalidRequest{Err: "[PerformMint] mint token null mint"}
	}

	// Assuming you're working within a function:
	mintAmount, err := strconv.ParseUint(bindingMsg.MintAmount, 10, 64)
	if err != nil {
		return errorsmod.Wrap(err, "[PerformMint] failed to parse redumption mint amount.")
	}

	if mintAmount > math.MaxInt64 {
		return errorsmod.Wrap(errortypes.ErrInvalidMintAmount, "[PerformMint] failed. Mint amount exceeds maximum allowed value.")
	}

	msg := &pooltypes.MsgMintTribute{
		Creator:         bindingMsg.Creator,
		ContractAddress: bindingMsg.ContractAddress,
		MintAmount:      mintAmount,
		ReceiptAddress:  bindingMsg.ReceiptAddress,
	}

	// Mint through token min / message server
	msgServer := poolkeeper.NewMsgServerImpl(*f)
	_, err = msgServer.MintTribute(ctx, msg)
	if err != nil {
		return errorsmod.Wrap(err, "[PerformMint] failed to mint native token.")
	}
	return nil
}
