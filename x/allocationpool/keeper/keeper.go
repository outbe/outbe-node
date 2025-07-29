package keeper

import (
	"fmt"
	"time"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/outbe/outbe-node/x/allocationpool/types"
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
		mintKeeper       types.MintKeeper
		feeCollectorName string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,

	stakingKeeper types.StakingKeeper,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	mintKeeper types.MintKeeper,
	feeCollectorName string,
) Keeper {
	// ensure mint module account is set
	// if addr := accountKeeper.GetModuleAddress(types.ModuleName); addr == nil {
	// 	sdkerrors.Wrap(errortypes.ErrKeyNotFound, "[GetModuleAddress] failed. The mint module account has not been set.")
	// 	return Keeper{}
	// }

	return Keeper{

		cdc:          cdc,
		storeService: storeService,

		stakingKeeper:    stakingKeeper,
		accountKeeper:    accountKeeper,
		bankKeeper:       bankKeeper,
		mintKeeper:       mintKeeper,
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
