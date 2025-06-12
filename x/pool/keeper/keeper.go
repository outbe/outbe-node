package keeper

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/outbe/outbe-node/x/pool/types"
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
		feeCollectorName string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,

	stakingKeeper types.StakingKeeper,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
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
		feeCollectorName: feeCollectorName,
	}
}

// GetLogger returns a logger instance with optional log printing based on the PrintLogs environment variable.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	printLogs, err := strconv.ParseBool(os.Getenv("PrintLogs"))
	if err != nil {
		fmt.Println("[Keeper][GetLogger] Error parsing PrintLogs environment variable:", err)
	}

	if !printLogs {
		return log.NewNopLogger()
	}

	logger := ctx.Logger().With(
		"timestamp", time.Now().UTC().Format(time.RFC3339),
		"module", fmt.Sprintf("x/%s", types.ModuleName),
		"height", ctx.BlockHeight(),
	)

	return logger
}
