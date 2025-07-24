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
	"github.com/outbe/outbe-node/x/allocationpool/constants"
	"github.com/outbe/outbe-node/x/allocationpool/types"
	mintType "github.com/outbe/outbe-node/x/gemmint/types"
)

func (k msgServer) MintTribute(goCtx context.Context, msg *types.MsgMintTribute) (*types.MsgMintTributeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	logger := k.Logger(ctx)

	log.Println("########## Mint Tribute Transaction Started ##########")

	checkEligibleContract := k.mintKeeper.GetWhitelist(ctx)

	if len(checkEligibleContract) == 0 {
		return nil, sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[MintTribute][GetWhitelist] failed. No smart contract is registered. Register an eligible smart contract first to be able mint tribute for it.")
	}

	if !k.mintKeeper.IsEligibleSmartContract(ctx, msg.ContractAddress) {
		return nil, sdkerrors.Wrap(errortypes.ErrUnauthorized, "[MintTribute][IsEligibleSmartContract] failed. No smart contract is registered. Given contract address is not eligible to mint.")
	}

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

	k.Logger(ctx).Info("[MintTribute] Retrieved total supply", "supply", totalSupply)

	if len(totalSupply) == 0 || totalSupply[0].TotalSupply == "" {
		currentSupply = 0
		strTotalSupply = "0"
		k.Logger(ctx).Info("[MintTribute] First mint", "currentSupply", currentSupply, "strTotalSupply", strTotalSupply)
	} else {
		var err error
		currentSupply, err = strconv.ParseUint(totalSupply[0].TotalSupply, 10, 64)
		if err != nil {
			return nil, sdkerrors.Wrap(errortypes.ErrInvalidType, "[MintTribute][ParseUint] failed. failed to parse total supply.")
		}
		k.Logger(ctx).Info("[MintTribute] Second mint", "currentSupply", currentSupply)
	}

	k.Logger(ctx).Info("[MintTribute] Before minting", "current_supply", currentSupply, "mint_amount", msg.MintAmount)

	currentSupply += msg.MintAmount

	k.Logger(ctx).Info("[MintTribute] Total minted before saving", "total_mint_amount", currentSupply)

	// Convert back to string
	strTotalSupply = strconv.FormatUint(currentSupply, 10)

	// Create supply object
	supply := types.Supply{
		TotalSupply: strTotalSupply,
	}
	if err := k.SetSupply(ctx, supply); err != nil {
		return nil, sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[MintTribute][SetSupply] failed. Couldn't store supply into the chain.")
	}

	k.Logger(ctx).Info("[MintTribute] Prepared supply for storage", "supply", supply)

	minted := mintType.Minted{
		TotalMinted: sdkmath.LegacyMustNewDecFromStr(strTotalSupply),
	}

	if err := k.mintKeeper.SetTotalMinted(ctx, minted); err != nil {
		return nil, sdkerrors.Wrap(errortypes.ErrInvalidRequest, "[MintTribute][SetTotalMinted] failed. Couldn't store total minted into the chain.")
	}

	k.Logger(ctx).Info("[MintTribute] Prepared minted for storage", "total_minted", minted)

	minted1, _ := k.mintKeeper.GetTotalMinted(ctx)

	k.Logger(ctx).Info("[MintTribute] fetching minted", "total_minted", minted1)

	verifySupply := k.TotalSupplyAll(ctx)

	k.Logger(ctx).Info("[MintTribute] Verified saved supply", "verified_supply", verifySupply)

	if len(verifySupply) == 0 || verifySupply[0].TotalSupply != strTotalSupply {
		return nil, fmt.Errorf("failed to verify saved supply: expected %s, got %v", strTotalSupply, verifySupply)
	}

	k.Logger(ctx).Info("[MintTribute] Prepared minted for storage", "verify_supply", verifySupply)

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

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventType,
			sdk.NewAttribute(types.AttributeKeyMintAmount, mintCoin.String()),
			sdk.NewAttribute(types.AttributeKeyTotalEmission, emission.TotalEmission),
		),
	)

	logger.Info("########## Mint Tribute Transaction Transaction Completed ##########")

	return &types.MsgMintTributeResponse{}, nil
}

func (k msgServer) BlockProvisionAmount(ctx sdk.Context) (uint64, error) {

	if ctx.BlockHeight() < constants.TransitionBlockNumber {

		tokens, err := k.CalculateExponentialBlockEmission(ctx, ctx.BlockHeight())
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
