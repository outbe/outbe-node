package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/outbe/outbe-node/x/distribution/types"
)

type BaseKeeper struct {
	distributionkeeper.Keeper

	storeService  store.KVStoreService
	cdc           codec.BinaryCodec
	authKeeper    stakingtypes.AccountKeeper
	bankKeeper    stakingtypes.BankKeeper
	stakingKeeper stakingtypes.StakingKeeper
	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string

	Schema  collections.Schema
	Params  collections.Item[types.Params]
	FeePool collections.Item[types.FeePool]

	feeCollectorName string // name of the FeeCollector ModuleAccount
}

func NewKeeper(
	cdc codec.BinaryCodec, storeService store.KVStoreService,
	ak stakingtypes.AccountKeeper, bk stakingtypes.BankKeeper, sk stakingtypes.StakingKeeper,
	feeCollectorName, authority string,
) BaseKeeper {
	// ensure distribution module account is set
	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	sb := collections.NewSchemaBuilder(storeService)
	k := BaseKeeper{
		storeService:     storeService,
		cdc:              cdc,
		authKeeper:       ak,
		bankKeeper:       bk,
		stakingKeeper:    sk,
		feeCollectorName: feeCollectorName,
		authority:        authority,
		Params:           collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		FeePool:          collections.NewItem(sb, types.FeePoolKey, "fee_pool", codec.CollValue[types.FeePool](cdc)),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}
