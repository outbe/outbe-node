package keeper

import (
	addresscodec "cosmossdk.io/core/address"
	"cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/outbe/outbe-node/x/staking/types"
)

type BaseKeeper struct {
	stakingkeeper.Keeper

	storeService          store.KVStoreService
	cdc                   codec.BinaryCodec
	ak                    stakingtypes.AccountKeeper
	bk                    stakingtypes.BankKeeper
	authority             string
	validatorAddressCodec addresscodec.Codec
	consensusAddressCodec addresscodec.Codec
}

func NewBaseKeeper(
	storeService store.KVStoreService,
	cdc codec.BinaryCodec,
	ak stakingtypes.AccountKeeper,
	bk stakingtypes.BankKeeper,
	authority string,
	validatorAddressCodec addresscodec.Codec,
	consensusAddressCodec addresscodec.Codec,
) BaseKeeper {
	// Initialize the base staking keeper using the proper constructor
	baseKeeper := stakingkeeper.NewKeeper(
		cdc,
		storeService,
		ak,
		bk,
		authority,
		validatorAddressCodec,
		consensusAddressCodec,
	)

	return BaseKeeper{
		Keeper:                *baseKeeper,
		storeService:          storeService,
		cdc:                   cdc,
		ak:                    ak,
		bk:                    bk,
		authority:             authority,
		validatorAddressCodec: validatorAddressCodec,
		consensusAddressCodec: consensusAddressCodec,
	}
}
