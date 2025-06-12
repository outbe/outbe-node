package keeper

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"

	errortypes "github.com/outbe/outbe-node/errors"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/outbe/outbe-node/app/params"
	"github.com/outbe/outbe-node/x/pool/constants"
	"github.com/outbe/outbe-node/x/pool/types"
)

func (k msgServer) MintTribute(goCtx context.Context, msg *types.MsgMintTribute) (*types.MsgMintTributeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	logger := k.Logger(ctx)

	log.Println("########## Mint Tribute Transaction Started ##########")

	if msg.MintAmount <= 0 {
		return nil, sdkerrors.Wrap(errortypes.ErrInvalidMintAmount, "[MintTribute] failed. Mint amount must be greater than zero")
	}

	emission, found := k.GetTotalEmission(ctx)
	if !found {
		return nil, sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[MintTribute][GetTotalEmission] failed. Total emission not found.")
	}

	decEmission, err := sdkmath.LegacyNewDecFromStr(emission.TotalEmission)
	if err != nil {
		return nil, sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[MintTribute][LegacyNewDecFromStr] failed. Total emission not converted.")
	}

	if msg.MintAmount > math.MaxInt64 {
		return nil, sdkerrors.Wrap(errortypes.ErrInvalidMintAmount, "[MintTribute] failed. Mint amount exceeds maximum allowed value.")
	}
	decMintAmount := sdkmath.LegacyNewDec(int64(msg.MintAmount))

	if decEmission.Sub(decMintAmount).LT(decMintAmount) {
		logger.Error("[MintTribute] Failed to mint coins. No emmisioned coin in pool for mint.",
			"mint_amount", msg.MintAmount,
		)
		return nil, sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[MintTribute][Sub] failed. No coin for mint.")
	}

	emission.TotalEmission = decEmission.Sub(decMintAmount).String()
	err = k.SetEmission(ctx, emission)
	if err != nil {
		return nil, sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[MintTribute][SetEmission] failed. Total emission not decreased.")
	}

	// Get current total supply
	totalSupply := k.TotalSupplyAll(ctx)

	var currentSupply uint64
	var strTotalSupply string

	// Log the retrieved supply for debugging
	k.Logger(ctx).Info("[MintTribute] Retrieved total supply", "supply", totalSupply)

	// Check if supply exists and is not empty
	if len(totalSupply) == 0 || totalSupply[0].TotalSupply == "" {
		currentSupply = 0
		strTotalSupply = "0"
	} else {
		// Parse existing supply
		var err error
		currentSupply, err = strconv.ParseUint(totalSupply[0].TotalSupply, 10, 64)
		if err != nil {
			return nil, sdkerrors.Wrap(errortypes.ErrInvalidType, "[MintTribute][ParseUint] failed. failed to parse total supply.")
		}
	}

	// Add mint amount
	k.Logger(ctx).Info("[MintTribute] Before minting", "current_supply", currentSupply, "mint_amount", msg.MintAmount)
	currentSupply += msg.MintAmount

	// Log the minting operation
	k.Logger(ctx).Info("[MintTribute] Total minted before saving", "total_mint_amount", currentSupply)

	// Convert back to string
	strTotalSupply = strconv.FormatUint(currentSupply, 10)

	// Create supply object
	supply := types.Supply{
		TotalSupply: strTotalSupply,
	}

	k.Logger(ctx).Info("[MintTribute] Prepared supply for storage", "supply", supply)

	// Save to state
	if err := k.SetSupply(ctx, supply); err != nil {
		return nil, sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[MintTribute][SetSupply] failed. Couldn't store supply into the chain.")
	}

	// Verify the save by querying immediately
	verifySupply := k.TotalSupplyAll(ctx)
	k.Logger(ctx).Info("[MintTribute] Verified saved supply", "verified_supply", verifySupply)

	if len(verifySupply) == 0 || verifySupply[0].TotalSupply != strTotalSupply {
		return nil, fmt.Errorf("failed to verify saved supply: expected %s, got %v", strTotalSupply, verifySupply)
	}

	// Mint the coins
	mintCoin := sdk.NewCoin(params.BondDenom, sdkmath.NewInt(int64(msg.MintAmount)))
	mintCoins := sdk.NewCoins(mintCoin)
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, mintCoins); err != nil {
		logger.Error("[MintTribute][MintCoins] Failed to mint coins",
			"mint_amount", msg.MintAmount,
			"error", err,
		)
		return nil, sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[MintTribute][MintCoins] failed. Failed to mint coins")
	}

	recipientAddress, err := sdk.AccAddressFromBech32(msg.ReceiptAddress)
	if err != nil {
		logger.Error("[MintTribute][AccAddressFromBech32] Invalid recipient address",
			"receipt_address", msg.ReceiptAddress,
			"error", err,
		)
		return nil, errors.New("[MintTribute][AccAddressFromBech32] failed. Invalid recipient address")
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		recipientAddress,
		sdk.Coins{mintCoin},
	)
	if err != nil {
		logger.Error("[MintTribute][SendCoinsFromModuleToAccount] failed to send coins",
			"receipt_address", msg.ReceiptAddress,
			"mint_amount", msg.MintAmount,
			"error", err,
		)

		return nil, errors.New("[MintTribute][SendCoinsFromModuleToAccount] failed to send coins to recipient")
	}

	id, err := k.GenerateTributeID(ctx)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "[MintTribute] failed to generate tribute ID")
	}

	// Create and store new tribute
	newTribute := types.Tribute{
		Id:               id,
		Creator:          msg.Creator,
		ContractAddress:  msg.ContractAddress,
		RecipientAddress: msg.ReceiptAddress,
		Amount:           msg.MintAmount,
	}
	if err := k.SetTribute(ctx, newTribute); err != nil {
		return nil, sdkerrors.Wrap(err, "[MintTribute] failed to store tribute")
	}

	logger.Info("########## Mint Tribute Transaction Transaction Completed ##########")

	return &types.MsgMintTributeResponse{}, nil
}

func (k msgServer) BlockProvisionAmount(ctx sdk.Context) (uint64, error) {

	if ctx.BlockHeight() < constants.TransitionBlockNumber {

		tokens, err := k.CalculateExponentialBlockEmission(ctx.BlockHeight())
		if err != nil {
			return 0, sdkerrors.Wrapf(errortypes.ErrInvalidRequest, "[BlockProvisionAmount][CalculateExponentialTokens] failed. ")
		}
		val, _ := strconv.ParseUint(tokens, 10, 64)
		return val, nil
	}

	tokens, err := k.CalculateFixedBlockEmission(ctx)
	if err != nil {
		return 0, errors.New("CalculateFixedTokens failed")
	}
	val, _ := strconv.ParseUint(tokens, 10, 64)
	return val, nil
}
