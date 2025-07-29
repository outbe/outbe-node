package keeper

import (
	"context"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	errortypes "github.com/outbe/outbe-node/errors"
)

type msgServer struct {
	WrappedBaseKeeper
}

func NewMsgServerImpl(keeper WrappedBaseKeeper) types.MsgServer {
	return &msgServer{WrappedBaseKeeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) WithdrawDelegatorReward(ctx context.Context, msg *types.MsgWithdrawDelegatorReward) (*types.MsgWithdrawDelegatorRewardResponse, error) {
	valAddr, err := k.sk.ValidatorAddressCodec().StringToBytes(msg.ValidatorAddress)
	if err != nil {
		return nil, errortypes.ErrNoValidatorForAddress.Wrapf("invalid validator address: %s", err)
	}

	delegatorAddress, err := k.ak.AddressCodec().StringToBytes(msg.DelegatorAddress)
	if err != nil {
		return nil, errortypes.ErrNoValidatorForAddress.Wrapf("invalid delegator address: %s", err)
	}

	amount, err := k.WithdrawDelegationRewards(ctx, delegatorAddress, valAddr)
	if err != nil {
		return nil, err
	}

	return &types.MsgWithdrawDelegatorRewardResponse{Amount: amount}, nil
}

func (k msgServer) SetWithdrawAddress(ctx context.Context, msg *types.MsgSetWithdrawAddress) (*types.MsgSetWithdrawAddressResponse, error) {
	delegatorAddress, err := k.ak.AddressCodec().StringToBytes(msg.DelegatorAddress)
	if err != nil {
		return nil, errortypes.ErrNoValidatorForAddress.Wrapf("invalid delegator address: %s", err)
	}

	withdrawAddress, err := k.ak.AddressCodec().StringToBytes(msg.WithdrawAddress)
	if err != nil {
		return nil, errortypes.ErrNoValidatorForAddress.Wrapf("invalid withdraw address: %s", err)
	}

	err = k.SetWithdrawAddr(ctx, delegatorAddress, withdrawAddress)
	if err != nil {
		return nil, err
	}

	return &types.MsgSetWithdrawAddressResponse{}, nil
}

func (k msgServer) WithdrawValidatorCommission(ctx context.Context, msg *types.MsgWithdrawValidatorCommission) (*types.MsgWithdrawValidatorCommissionResponse, error) {
	valAddr, err := k.sk.ValidatorAddressCodec().StringToBytes(msg.ValidatorAddress)
	if err != nil {
		return nil, errortypes.ErrNoValidatorForAddress.Wrapf("invalid validator address: %s", err)
	}

	amount, err := k.Keeper.WithdrawValidatorCommission(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	return &types.MsgWithdrawValidatorCommissionResponse{Amount: amount}, nil
}

func (k msgServer) FundCommunityPool(ctx context.Context, msg *types.MsgFundCommunityPool) (*types.MsgFundCommunityPoolResponse, error) {
	depositor, err := k.ak.AddressCodec().StringToBytes(msg.Depositor)
	if err != nil {
		return nil, errortypes.ErrNoValidatorForAddress.Wrapf("invalid depositor address: %s", err)
	}

	if err := k.Keeper.FundCommunityPool(ctx, msg.Amount, depositor); err != nil {
		return nil, err
	}

	return &types.MsgFundCommunityPoolResponse{}, nil
}

func (k msgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {

	if err := msg.Params.ValidateBasic(); err != nil {
		return nil, err
	}

	if err := k.Params.Set(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

func (k msgServer) CommunityPoolSpend(ctx context.Context, msg *types.MsgCommunityPoolSpend) (*types.MsgCommunityPoolSpendResponse, error) {

	recipient, err := k.ak.AddressCodec().StringToBytes(msg.Recipient)
	if err != nil {
		return nil, err
	}

	if err := k.DistributeFromFeePool(ctx, msg.Amount, recipient); err != nil {
		return nil, err
	}

	logger := k.Logger(ctx)
	logger.Info("transferred from the community pool to recipient", "amount", msg.Amount.String(), "recipient", msg.Recipient)

	return &types.MsgCommunityPoolSpendResponse{}, nil
}

func (k msgServer) DepositValidatorRewardsPool(ctx context.Context, msg *types.MsgDepositValidatorRewardsPool) (*types.MsgDepositValidatorRewardsPoolResponse, error) {

	valAddr, err := k.sk.ValidatorAddressCodec().StringToBytes(msg.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	validator, err := k.sk.Validator(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	if validator == nil {
		return nil, errors.Wrapf(types.ErrNoValidatorExists, msg.ValidatorAddress)
	}

	// Allocate tokens from the distribution module to the validator, which are
	// then distributed to the validator's delegators.
	reward := sdk.NewDecCoinsFromCoins(msg.Amount...)
	if err = k.AllocateTokensToValidator(ctx, validator, reward); err != nil {
		return nil, err
	}

	logger := k.Logger(ctx)
	logger.Info(
		"transferred from rewards to validator rewards pool",
		"depositor", msg.Depositor,
		"amount", msg.Amount.String(),
		"validator", msg.ValidatorAddress,
	)

	return &types.MsgDepositValidatorRewardsPoolResponse{}, nil
}
