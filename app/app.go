package app

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"sync"

	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/gogoproto/proto"
	"github.com/cosmos/ibc-go/modules/capability"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	ica "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts"
	icacontroller "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	icahost "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	ibcfee "github.com/cosmos/ibc-go/v8/modules/apps/29-fee"
	ibcfeekeeper "github.com/cosmos/ibc-go/v8/modules/apps/29-fee/keeper"
	ibcfeetypes "github.com/cosmos/ibc-go/v8/modules/apps/29-fee/types"

	"github.com/cosmos/ibc-go/v8/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v8/modules/core"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types" //nolint:staticcheck
	ibcconnectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	ibctm "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	"github.com/spf13/cast"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/circuit"
	circuitkeeper "cosmossdk.io/x/circuit/keeper"
	circuittypes "cosmossdk.io/x/circuit/types"
	"cosmossdk.io/x/evidence"
	evidencekeeper "cosmossdk.io/x/evidence/keeper"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	feegrantmodule "cosmossdk.io/x/feegrant/module"
	"cosmossdk.io/x/nft"
	nftkeeper "cosmossdk.io/x/nft/keeper"
	nftmodule "cosmossdk.io/x/nft/module"
	"cosmossdk.io/x/upgrade"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"

	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	signingtype "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/posthandler"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	txmodule "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"

	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/group"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"
	groupmodule "github.com/cosmos/cosmos-sdk/x/group/module"
	"github.com/cosmos/cosmos-sdk/x/mint"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"

	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvm "github.com/CosmWasm/wasmvm/v2"

	wasmlc "github.com/cosmos/ibc-go/modules/light-clients/08-wasm"
	wasmlckeeper "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/keeper"
	wasmlctypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/types"

	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward"
	packetforwardkeeper "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/keeper"
	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/types"

	ratelimit "github.com/cosmos/ibc-apps/modules/rate-limiting/v8"
	ratelimitkeeper "github.com/cosmos/ibc-apps/modules/rate-limiting/v8/keeper"
	ratelimittypes "github.com/cosmos/ibc-apps/modules/rate-limiting/v8/types"

	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/staking"

	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	chainante "github.com/outbe/outbe-node/app/ante"
)

const (
	appName      = "outbe-node"
	NodeDir      = ".outbe-node"
	Bech32Prefix = "outbe"

	ChainID = "localchain-1"
)

var (
	capabilities = []string{
		"iterator",
		"staking",
		"stargate",
		"cosmwasm_1_1",
		"cosmwasm_1_2",
		"cosmwasm_1_3",
		"cosmwasm_1_4",
		"cosmwasm_2_0",
		"cosmwasm_2_1",
		"cosmwasm_2_2",
	}
)

func init() {
	// manually update the power reduction based on the base denom unit (10^18 [evm] or 10^6 [cosmos])
	sdk.DefaultPowerReduction = math.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(BaseDenomUnit), nil))
}

// These constants are derived from the above variables.
// These are the ones we will want to use in the code, based on
// any overrides above
var (
	// DefaultNodeHome default home directories for appd
	DefaultNodeHome = os.ExpandEnv("$HOME/") + NodeDir

	CoinType uint32 = 118

	BaseDenomUnit int64 = 18

	BaseDenom    = "unit"
	DisplayDenom = "UNIT"

	// Bech32PrefixAccAddr defines the Bech32 prefix of an account's address
	Bech32PrefixAccAddr = Bech32Prefix
	// Bech32PrefixAccPub defines the Bech32 prefix of an account's public key
	Bech32PrefixAccPub = Bech32Prefix + sdk.PrefixPublic
	// Bech32PrefixValAddr defines the Bech32 prefix of a validator's operator address
	Bech32PrefixValAddr = Bech32Prefix + sdk.PrefixValidator + sdk.PrefixOperator
	// Bech32PrefixValPub defines the Bech32 prefix of a validator's operator public key
	Bech32PrefixValPub = Bech32Prefix + sdk.PrefixValidator + sdk.PrefixOperator + sdk.PrefixPublic
	// Bech32PrefixConsAddr defines the Bech32 prefix of a consensus node address
	Bech32PrefixConsAddr = Bech32Prefix + sdk.PrefixValidator + sdk.PrefixConsensus
	// Bech32PrefixConsPub defines the Bech32 prefix of a consensus node public key
	Bech32PrefixConsPub = Bech32Prefix + sdk.PrefixValidator + sdk.PrefixConsensus + sdk.PrefixPublic
)

// module account permissions
var maccPerms = map[string][]string{
	authtypes.FeeCollectorName:     nil,
	distrtypes.ModuleName:          nil,
	minttypes.ModuleName:           {authtypes.Minter},
	stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
	stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
	govtypes.ModuleName:            {authtypes.Burner},
	nft.ModuleName:                 nil,
	// non sdk modules
	ibctransfertypes.ModuleName: {authtypes.Minter, authtypes.Burner},
	ibcfeetypes.ModuleName:      nil,
	icatypes.ModuleName:         nil,
	wasmtypes.ModuleName:        {authtypes.Burner},
}

var (
	_ runtime.AppI            = (*ChainApp)(nil)
	_ servertypes.Application = (*ChainApp)(nil)
)

// ChainApp extended ABCI application
type ChainApp struct {
	*baseapp.BaseApp
	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry types.InterfaceRegistry

	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	// keepers
	AccountKeeper         authkeeper.AccountKeeper
	BankKeeper            bankkeeper.BaseKeeper
	CapabilityKeeper      *capabilitykeeper.Keeper
	StakingKeeper         *stakingkeeper.Keeper
	SlashingKeeper        slashingkeeper.Keeper
	MintKeeper            mintkeeper.Keeper
	DistrKeeper           distrkeeper.Keeper
	GovKeeper             govkeeper.Keeper
	CrisisKeeper          *crisiskeeper.Keeper
	UpgradeKeeper         *upgradekeeper.Keeper
	ParamsKeeper          paramskeeper.Keeper
	AuthzKeeper           authzkeeper.Keeper
	EvidenceKeeper        evidencekeeper.Keeper
	FeeGrantKeeper        feegrantkeeper.Keeper
	GroupKeeper           groupkeeper.Keeper
	NFTKeeper             nftkeeper.Keeper
	ConsensusParamsKeeper consensusparamkeeper.Keeper
	CircuitKeeper         circuitkeeper.Keeper

	IBCKeeper           *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	IBCFeeKeeper        ibcfeekeeper.Keeper
	ICAControllerKeeper icacontrollerkeeper.Keeper
	ICAHostKeeper       icahostkeeper.Keeper
	TransferKeeper      ibctransferkeeper.Keeper

	// Custom
	WasmKeeper          wasmkeeper.Keeper
	PacketForwardKeeper *packetforwardkeeper.Keeper
	WasmClientKeeper    wasmlckeeper.Keeper
	RatelimitKeeper     ratelimitkeeper.Keeper

	ScopedIBCKeeper           capabilitykeeper.ScopedKeeper
	ScopedICAHostKeeper       capabilitykeeper.ScopedKeeper
	ScopedICAControllerKeeper capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper      capabilitykeeper.ScopedKeeper
	ScopedIBCFeeKeeper        capabilitykeeper.ScopedKeeper
	ScopedWasmKeeper          capabilitykeeper.ScopedKeeper

	// the module manager
	ModuleManager      *module.Manager
	BasicModuleManager module.BasicManager

	// simulation manager
	sm *module.SimulationManager

	// module configurator
	configurator module.Configurator
	once         sync.Once
}

// NewChainApp returns a reference to an initialized ChainApp.
func NewChainApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	wasmOpts []wasmkeeper.Option,
	baseAppOptions ...func(*baseapp.BaseApp),
) *ChainApp {

	interfaceRegistry := GetInterfaceRegistry()
	appCodec := codec.NewProtoCodec(interfaceRegistry)
	legacyAmino := codec.NewLegacyAmino()
	txConfig := authtx.NewTxConfig(appCodec, authtx.DefaultSignModes)
	std.RegisterLegacyAminoCodec(legacyAmino)
	std.RegisterInterfaces(interfaceRegistry)

	// Below we could construct and set an application specific mempool and
	// ABCI 1.0 PrepareProposal and ProcessProposal handlers. These defaults are
	// already set in the SDK's BaseApp, this shows an example of how to override
	// them.
	//
	// Example:
	//
	// bApp := baseapp.NewBaseApp(...)
	// nonceMempool := mempool.NewSenderNonceMempool()
	// abciPropHandler := NewDefaultProposalHandler(nonceMempool, bApp)
	//
	// bApp.SetMempool(nonceMempool)
	// bApp.SetPrepareProposal(abciPropHandler.PrepareProposalHandler())
	// bApp.SetProcessProposal(abciPropHandler.ProcessProposalHandler())
	//
	// Alternatively, you can construct BaseApp options, append those to
	// baseAppOptions and pass them to NewBaseApp.
	//
	// Example:
	//
	// prepareOpt = func(app *baseapp.BaseApp) {
	// 	abciPropHandler := baseapp.NewDefaultProposalHandler(nonceMempool, app)
	// 	app.SetPrepareProposal(abciPropHandler.PrepareProposalHandler())
	// }
	// baseAppOptions = append(baseAppOptions, prepareOpt)

	// create and set dummy vote extension handler
	// voteExtOp := func(bApp *baseapp.BaseApp) {
	//	voteExtHandler := NewVoteExtensionHandler()
	//	voteExtHandler.SetHandlers(bApp)
	// }
	// baseAppOptions = append(baseAppOptions, voteExtOp)

	baseAppOptions = append(baseAppOptions, baseapp.SetOptimisticExecution())

	bApp := baseapp.NewBaseApp(appName, logger, db, txConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)

	bApp.SetTxEncoder(txConfig.TxEncoder())

	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey,
		stakingtypes.StoreKey,
		crisistypes.StoreKey,
		minttypes.StoreKey,
		distrtypes.StoreKey,
		slashingtypes.StoreKey,
		govtypes.StoreKey,
		paramstypes.StoreKey,
		consensusparamtypes.StoreKey,
		upgradetypes.StoreKey,
		feegrant.StoreKey,
		evidencetypes.StoreKey,
		circuittypes.StoreKey,
		authzkeeper.StoreKey,
		nftkeeper.StoreKey,
		group.StoreKey,
		// non sdk store keys
		capabilitytypes.StoreKey,
		ibcexported.StoreKey,
		ibctransfertypes.StoreKey,
		ibcfeetypes.StoreKey,
		wasmtypes.StoreKey,
		icahosttypes.StoreKey,
		icacontrollertypes.StoreKey,
		packetforwardtypes.StoreKey,
		wasmlctypes.StoreKey,
		ratelimittypes.StoreKey,
	)

	tkeys := storetypes.NewTransientStoreKeys(
		paramstypes.TStoreKey,
	)
	memKeys := storetypes.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)

	// register streaming services
	if err := bApp.RegisterStreamingServices(appOpts, keys); err != nil {
		panic(err)
	}

	app := &ChainApp{
		BaseApp:           bApp,
		legacyAmino:       legacyAmino,
		appCodec:          appCodec,
		txConfig:          txConfig,
		interfaceRegistry: interfaceRegistry,
		keys:              keys,
		tkeys:             tkeys,
		memKeys:           memKeys,
	}

	app.ParamsKeeper = initParamsKeeper(
		appCodec,
		legacyAmino,
		keys[paramstypes.StoreKey],
		tkeys[paramstypes.TStoreKey],
	)

	// set the BaseApp's parameter store
	app.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[consensusparamtypes.StoreKey]),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		runtime.EventService{},
	)
	bApp.SetParamStore(app.ConsensusParamsKeeper.ParamsStore)

	// add capability keeper and ScopeToModule for ibc module
	app.CapabilityKeeper = capabilitykeeper.NewKeeper(
		appCodec,
		keys[capabilitytypes.StoreKey],
		memKeys[capabilitytypes.MemStoreKey],
	)

	scopedIBCKeeper := app.CapabilityKeeper.ScopeToModule(ibcexported.ModuleName)
	scopedICAHostKeeper := app.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	scopedICAControllerKeeper := app.CapabilityKeeper.ScopeToModule(icacontrollertypes.SubModuleName)
	scopedTransferKeeper := app.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	scopedWasmKeeper := app.CapabilityKeeper.ScopeToModule(wasmtypes.ModuleName)
	app.CapabilityKeeper.Seal()

	// add keepers

	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		app.AccountKeeper,
		BlockedAddresses(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		logger,
	)

	// enable sign mode textual by overwriting the default tx config (after setting the bank keeper)
	enabledSignModes := append(tx.DefaultSignModes, signingtype.SignMode_SIGN_MODE_TEXTUAL)
	txConfigOpts := tx.ConfigOptions{
		EnabledSignModes:           enabledSignModes,
		TextualCoinMetadataQueryFn: txmodule.NewBankKeeperCoinMetadataQueryFn(app.BankKeeper),
	}
	txConfig, err := tx.NewTxConfigWithOptions(
		appCodec,
		txConfigOpts,
	)
	if err != nil {
		panic(err)
	}
	app.txConfig = txConfig

	app.StakingKeeper = stakingkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[stakingtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	)
	app.MintKeeper = mintkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[minttypes.StoreKey]),
		app.StakingKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.DistrKeeper = distrkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[distrtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec,
		legacyAmino,
		runtime.NewKVStoreService(keys[slashingtypes.StoreKey]),
		app.StakingKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	invCheckPeriod := cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod))
	app.CrisisKeeper = crisiskeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[crisistypes.StoreKey]),
		invCheckPeriod,
		app.BankKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		app.AccountKeeper.AddressCodec(),
	)

	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(appCodec, runtime.NewKVStoreService(keys[feegrant.StoreKey]), app.AccountKeeper)

	app.CircuitKeeper = circuitkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[circuittypes.StoreKey]),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		app.AccountKeeper.AddressCodec(),
	)
	app.BaseApp.SetCircuitBreaker(&app.CircuitKeeper)

	app.AuthzKeeper = authzkeeper.NewKeeper(
		runtime.NewKVStoreService(keys[authzkeeper.StoreKey]),
		appCodec,
		app.MsgServiceRouter(),
		app.AccountKeeper,
	)

	groupConfig := group.DefaultConfig()
	groupConfig.MaxMetadataLen = 10000
	app.GroupKeeper = groupkeeper.NewKeeper(
		keys[group.StoreKey],
		// runtime.NewKVStoreService(keys[group.StoreKey]),
		appCodec,
		app.MsgServiceRouter(),
		app.AccountKeeper,
		groupConfig,
	)

	// get skipUpgradeHeights from the app options
	skipUpgradeHeights := map[int64]bool{}
	for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}
	homePath := cast.ToString(appOpts.Get(flags.FlagHome))
	// set the governance module account as the authority for conducting upgrades
	app.UpgradeKeeper = upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		runtime.NewKVStoreService(keys[upgradetypes.StoreKey]),
		appCodec,
		homePath,
		app.BaseApp,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.IBCKeeper = ibckeeper.NewKeeper(
		appCodec,
		keys[ibcexported.StoreKey],
		app.GetSubspace(ibcexported.ModuleName),
		app.StakingKeeper,
		app.UpgradeKeeper,
		scopedIBCKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	app.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(
			app.DistrKeeper.Hooks(),
			app.SlashingKeeper.Hooks(),
		),
	)

	// Register the proposal types
	// Deprecated: Avoid adding new handlers, instead use the new proposal flow
	// by granting the governance module the right to execute the message.
	// See: https://docs.cosmos.network/main/modules/gov#proposal-messages
	govRouter := govv1beta1.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govv1beta1.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(app.ParamsKeeper))
	govConfig := govtypes.DefaultConfig()
	govConfig.MaxMetadataLen = 20000
	govKeeper := govkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[govtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		app.DistrKeeper,
		app.MsgServiceRouter(),
		govConfig,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Set legacy router for backwards compatibility with gov v1beta1
	govKeeper.SetLegacyRouter(govRouter)

	app.GovKeeper = *govKeeper.SetHooks(
		govtypes.NewMultiGovHooks(
		// register the governance hooks
		),
	)

	app.NFTKeeper = nftkeeper.NewKeeper(
		runtime.NewKVStoreService(keys[nftkeeper.StoreKey]),
		appCodec,
		app.AccountKeeper,
		app.BankKeeper,
	)

	// create evidence keeper with router
	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[evidencetypes.StoreKey]),
		app.StakingKeeper,
		app.SlashingKeeper,
		app.AccountKeeper.AddressCodec(),
		runtime.ProvideCometInfoService(),
	)
	// If evidence needs to be handled for the app, set routes in router here and seal
	app.EvidenceKeeper = *evidenceKeeper

	// IBC Fee Module keeper
	app.IBCFeeKeeper = ibcfeekeeper.NewKeeper(
		appCodec, keys[ibcfeetypes.StoreKey],
		app.IBCKeeper.ChannelKeeper, // may be replaced with IBC middleware
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper, app.AccountKeeper, app.BankKeeper,
	)

	// Create the ratelimit keeper
	app.RatelimitKeeper = *ratelimitkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[ratelimittypes.StoreKey]),
		app.GetSubspace(ratelimittypes.ModuleName),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		app.BankKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.IBCFeeKeeper, // ICS4Wrapper
	)

	// Create Transfer Keepers
	app.TransferKeeper = ibctransferkeeper.NewKeeper(
		appCodec,
		keys[ibctransfertypes.StoreKey],
		app.GetSubspace(ibctransfertypes.ModuleName),
		app.RatelimitKeeper, // ICS4Wrapper
		//app.IBCFeeKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		scopedTransferKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Create the packetfoward keeper
	app.PacketForwardKeeper = packetforwardkeeper.NewKeeper(
		appCodec,
		keys[packetforwardtypes.StoreKey],
		app.TransferKeeper, // will be zero-value here, reference is set later on with SetTransferKeeper.
		app.IBCKeeper.ChannelKeeper,
		app.BankKeeper,
		app.IBCKeeper.ChannelKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.PacketForwardKeeper.SetTransferKeeper(app.TransferKeeper)

	app.ICAHostKeeper = icahostkeeper.NewKeeper(
		appCodec,
		keys[icahosttypes.StoreKey],
		app.GetSubspace(icahosttypes.SubModuleName),
		app.IBCFeeKeeper, // use ics29 fee as ics4Wrapper in middleware stack
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		scopedICAHostKeeper,
		app.MsgServiceRouter(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.ICAHostKeeper.WithQueryRouter(app.GRPCQueryRouter())

	app.ICAControllerKeeper = icacontrollerkeeper.NewKeeper(
		appCodec,
		keys[icacontrollertypes.StoreKey],
		app.GetSubspace(icacontrollertypes.SubModuleName),
		app.IBCFeeKeeper, // use ics29 fee as ics4Wrapper in middleware stack
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper,
		scopedICAControllerKeeper,
		app.MsgServiceRouter(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	wasmOpts = append(wasmOpts)
	wasmDir := filepath.Join(homePath, "data")
	wasmConfig, err := wasm.ReadWasmConfig(appOpts)
	if err != nil {
		panic(fmt.Sprintf("error while reading wasm config: %s", err))
	}

	// The last arguments can contain custom message handlers, and custom query handlers,
	// if we want to allow any custom callbacks
	app.WasmKeeper = wasmkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[wasmtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		distrkeeper.NewQuerier(app.DistrKeeper),
		app.IBCFeeKeeper, // ISC4 Wrapper: fee IBC middleware
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper,
		scopedWasmKeeper,
		app.TransferKeeper,
		app.MsgServiceRouter(),
		app.GRPCQueryRouter(),
		wasmDir,
		wasmConfig,
		capabilities,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		wasmOpts...,
	)

	wasmLightClientQuerier := wasmlctypes.QueryPlugins{
		// Custom: MyCustomQueryPlugin(),
		// `myAcceptList` is a `[]string` containing the list of gRPC query paths that the chain wants to allow for the `08-wasm` module to query.
		// These queries must be registered in the chain's gRPC query router, be deterministic, and track their gas usage.
		// The `AcceptListStargateQuerier` function will return a query plugin that will only allow queries for the paths in the `myAcceptList`.
		// The query responses are encoded in protobuf unlike the implementation in `x/wasm`.
		Stargate: wasmlctypes.AcceptListStargateQuerier([]string{
			"/ibc.core.client.v1.Query/ClientState",
			"/ibc.core.client.v1.Query/ConsensusState",
			"/ibc.core.connection.v1.Query/Connection",
		}),
	}

	dataDir := filepath.Join(homePath, "data")

	var memCacheSizeMB uint32 = 100
	lc08, err := wasmvm.NewVM(filepath.Join(dataDir, "08-light-client"), capabilities, 32, false, memCacheSizeMB)
	if err != nil {
		panic(fmt.Sprintf("failed to create VM for 08 light client: %s", err))
	}

	app.WasmClientKeeper = wasmlckeeper.NewKeeperWithVM(
		appCodec,
		runtime.NewKVStoreService(keys[wasmlctypes.StoreKey]),
		app.IBCKeeper.ClientKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		lc08,
		bApp.GRPCQueryRouter(),
		wasmlckeeper.WithQueryPlugins(&wasmLightClientQuerier),
	)

	// Create Transfer Stack
	var transferStack porttypes.IBCModule
	transferStack = transfer.NewIBCModule(app.TransferKeeper)
	transferStack = ratelimit.NewIBCMiddleware(app.RatelimitKeeper, transferStack)
	transferStack = ibcfee.NewIBCMiddleware(transferStack, app.IBCFeeKeeper)
	transferStack = packetforward.NewIBCMiddleware(
		transferStack,
		app.PacketForwardKeeper,
		0,
		packetforwardkeeper.DefaultForwardTransferPacketTimeoutTimestamp,
	)

	// Create Interchain Accounts Stack
	// SendPacket, since it is originating from the application to core IBC:
	// icaAuthModuleKeeper.SendTx -> icaController.SendPacket -> fee.SendPacket -> channel.SendPacket
	var icaControllerStack porttypes.IBCModule
	// integration point for custom authentication modules
	// see https://medium.com/the-interchain-foundation/ibc-go-v6-changes-to-interchain-accounts-and-how-it-impacts-your-chain-806c185300d7
	var noAuthzModule porttypes.IBCModule
	icaControllerStack = icacontroller.NewIBCMiddleware(noAuthzModule, app.ICAControllerKeeper)
	icaControllerStack = ibcfee.NewIBCMiddleware(icaControllerStack, app.IBCFeeKeeper)

	// RecvPacket, message that originates from core IBC and goes down to app, the flow is:
	// channel.RecvPacket -> fee.OnRecvPacket -> icaHost.OnRecvPacket
	var icaHostStack porttypes.IBCModule
	icaHostStack = icahost.NewIBCModule(app.ICAHostKeeper)
	icaHostStack = ibcfee.NewIBCMiddleware(icaHostStack, app.IBCFeeKeeper)

	var wasmStack porttypes.IBCModule // Create fee enabled wasm ibc Stack
	wasmStack = wasm.NewIBCHandler(app.WasmKeeper, app.IBCKeeper.ChannelKeeper, app.IBCFeeKeeper)
	wasmStack = ibcfee.NewIBCMiddleware(wasmStack, app.IBCFeeKeeper)

	// Create static IBC router, add app routes, then set and seal it
	ibcRouter := porttypes.NewRouter()
	ibcRouter.AddRoute(ibctransfertypes.ModuleName, transferStack)
	ibcRouter.AddRoute(wasmtypes.ModuleName, wasmStack)
	ibcRouter.AddRoute(icacontrollertypes.SubModuleName, icaControllerStack)
	ibcRouter.AddRoute(icahosttypes.SubModuleName, icaHostStack)
	app.IBCKeeper.SetRouter(ibcRouter)

	// --- Module Options ---

	// NOTE: we may consider parsing `appOpts` inside module constructors. For the moment
	// we prefer to be more strict in what arguments the modules expect.
	skipGenesisInvariants := cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.
	app.ModuleManager = module.NewManager(
		genutil.NewAppModule(
			app.AccountKeeper,
			app.StakingKeeper,
			app,
			txConfig,
		),
		auth.NewAppModule(appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts, app.GetSubspace(authtypes.ModuleName)),
		vesting.NewAppModule(app.AccountKeeper, app.BankKeeper),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		gov.NewAppModule(appCodec, &app.GovKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(govtypes.ModuleName)),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper, nil, app.GetSubspace(minttypes.ModuleName)),
		slashing.NewAppModule(
			appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper,
			app.StakingKeeper,
			app.GetSubspace(slashingtypes.ModuleName), app.interfaceRegistry),
		distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(distrtypes.ModuleName)),
		staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(stakingtypes.ModuleName)),
		upgrade.NewAppModule(app.UpgradeKeeper, app.AccountKeeper.AddressCodec()),
		evidence.NewAppModule(app.EvidenceKeeper),
		params.NewAppModule(app.ParamsKeeper),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		groupmodule.NewAppModule(appCodec, app.GroupKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		nftmodule.NewAppModule(appCodec, app.NFTKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		consensus.NewAppModule(appCodec, app.ConsensusParamsKeeper),
		circuit.NewAppModule(appCodec, app.CircuitKeeper),
		// non sdk modules
		capability.NewAppModule(appCodec, *app.CapabilityKeeper, false),
		wasm.NewAppModule(appCodec, &app.WasmKeeper,
			app.StakingKeeper,
			app.AccountKeeper, app.BankKeeper, app.MsgServiceRouter(), app.GetSubspace(wasmtypes.ModuleName),
		),

		ibc.NewAppModule(app.IBCKeeper),
		transfer.NewAppModule(app.TransferKeeper),
		ibcfee.NewAppModule(app.IBCFeeKeeper),
		ica.NewAppModule(&app.ICAControllerKeeper, &app.ICAHostKeeper),
		ibctm.NewAppModule(),
		crisis.NewAppModule(app.CrisisKeeper, skipGenesisInvariants, app.GetSubspace(crisistypes.ModuleName)),
		// custom
		packetforward.NewAppModule(app.PacketForwardKeeper, app.GetSubspace(packetforwardtypes.ModuleName)),
		wasmlc.NewAppModule(app.WasmClientKeeper),
		ratelimit.NewAppModule(appCodec, app.RatelimitKeeper),
	)

	// BasicModuleManager defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration and genesis verification.
	// By default it is composed of all the module from the module manager.
	// Additionally, app module basics can be overwritten by passing them as argument.
	app.BasicModuleManager = module.NewBasicManagerFromManager(
		app.ModuleManager,
		map[string]module.AppModuleBasic{
			genutiltypes.ModuleName: genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
		})
	app.BasicModuleManager.RegisterLegacyAminoCodec(legacyAmino)
	app.BasicModuleManager.RegisterInterfaces(interfaceRegistry)

	// NOTE: upgrade module is required to be prioritized
	app.ModuleManager.SetOrderPreBlockers(
		upgradetypes.ModuleName,
	)
	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	// NOTE: capability module's beginblocker must come before any modules using capabilities (e.g. IBC)
	app.ModuleManager.SetOrderBeginBlockers(
		minttypes.ModuleName,

		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		genutiltypes.ModuleName,
		authz.ModuleName,
		// additional non simd modules
		capabilitytypes.ModuleName,
		ibctransfertypes.ModuleName,
		ibcexported.ModuleName,
		icatypes.ModuleName,
		ibcfeetypes.ModuleName,
		wasmtypes.ModuleName,
		packetforwardtypes.ModuleName,
		wasmlctypes.ModuleName,
		ratelimittypes.ModuleName,
	)

	app.ModuleManager.SetOrderEndBlockers(
		crisistypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		genutiltypes.ModuleName,
		feegrant.ModuleName,
		group.ModuleName,
		capabilitytypes.ModuleName,
		ibctransfertypes.ModuleName,
		ibcexported.ModuleName,
		icatypes.ModuleName,
		ibcfeetypes.ModuleName,
		wasmtypes.ModuleName,
		packetforwardtypes.ModuleName,
		wasmlctypes.ModuleName,
		ratelimittypes.ModuleName,
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: The genutils module must also occur after auth so that it can access the params from auth.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	// NOTE: wasm module should be at the end as it can call other module functionality direct or via message dispatching during
	// genesis phase. For example bank transfer, auth account check, staking, ...
	genesisModuleOrder := []string{
		capabilitytypes.ModuleName,
		// simd modules
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		crisistypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		nft.ModuleName,
		group.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		consensusparamtypes.ModuleName,
		circuittypes.ModuleName,
		// additional non simd modules
		ibctransfertypes.ModuleName,
		ibcexported.ModuleName,
		icatypes.ModuleName,
		ibcfeetypes.ModuleName,
		wasmtypes.ModuleName, // wasm after ibc transfer
		packetforwardtypes.ModuleName,
		wasmlctypes.ModuleName,
		ratelimittypes.ModuleName,
	}
	app.ModuleManager.SetOrderInitGenesis(genesisModuleOrder...)
	app.ModuleManager.SetOrderExportGenesis(genesisModuleOrder...)

	// Uncomment if you want to set a custom migration order here.
	// app.ModuleManager.SetOrderMigrations(custom order)

	app.ModuleManager.RegisterInvariants(app.CrisisKeeper)
	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	err = app.ModuleManager.RegisterServices(app.configurator)
	if err != nil {
		panic(err)
	}

	// RegisterUpgradeHandlers is used for registering any on-chain upgrades.
	// Make sure it's called after `app.ModuleManager` and `app.configurator` are set.
	app.RegisterUpgradeHandlers()

	autocliv1.RegisterQueryServer(app.GRPCQueryRouter(), runtimeservices.NewAutoCLIQueryService(app.ModuleManager.Modules))

	reflectionSvc, err := runtimeservices.NewReflectionService()
	if err != nil {
		panic(err)
	}
	reflectionv1.RegisterReflectionServiceServer(app.GRPCQueryRouter(), reflectionSvc)

	// add test gRPC service for testing gRPC queries in isolation
	// testdata_pulsar.RegisterQueryServer(app.GRPCQueryRouter(), testdata_pulsar.QueryImpl{})

	// create the simulation manager and define the order of the modules for deterministic simulations
	//
	// NOTE: this is not required apps that don't use the simulator for fuzz testing
	// transactions
	overrideModules := map[string]module.AppModuleSimulation{
		authtypes.ModuleName: auth.NewAppModule(app.appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts, app.GetSubspace(authtypes.ModuleName)),
	}
	app.sm = module.NewSimulationManagerFromAppModules(app.ModuleManager.Modules, overrideModules)

	app.sm.RegisterStoreDecoders()

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)
	app.MountMemoryStores(memKeys)

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetPreBlocker(app.PreBlocker)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)

	app.setAnteHandler(chainante.HandlerOptions{
		Cdc:                   app.appCodec,
		AccountKeeper:         app.AccountKeeper,
		BankKeeper:            app.BankKeeper,
		FeegrantKeeper:        app.FeeGrantKeeper,
		SignModeHandler:       txConfig.SignModeHandler(),
		IBCKeeper:             app.IBCKeeper,
		WasmKeeper:            &app.WasmKeeper,
		WasmConfig:            &wasmConfig,
		TXCounterStoreService: runtime.NewKVStoreService(keys[wasmtypes.StoreKey]),
		CircuitKeeper:         &app.CircuitKeeper,
		SigGasConsumer:        authante.DefaultSigVerificationGasConsumer,
	})

	// must be before Loading version
	// requires the snapshot store to be created and registered as a BaseAppOption
	// see cmd/outbe-noded/root.go: 206 - 214 approx
	if manager := app.SnapshotManager(); manager != nil {
		err := manager.RegisterExtensions(
			wasmkeeper.NewWasmSnapshotter(app.CommitMultiStore(), &app.WasmKeeper),
			wasmlckeeper.NewWasmSnapshotter(app.CommitMultiStore(), &app.WasmClientKeeper),
		)
		if err != nil {
			panic(fmt.Errorf("failed to register snapshot extension: %s", err))
		}
	}

	app.ScopedIBCKeeper = scopedIBCKeeper
	app.ScopedTransferKeeper = scopedTransferKeeper
	app.ScopedWasmKeeper = scopedWasmKeeper
	app.ScopedICAHostKeeper = scopedICAHostKeeper
	app.ScopedICAControllerKeeper = scopedICAControllerKeeper

	// In v0.46, the SDK introduces _postHandlers_. PostHandlers are like
	// antehandlers, but are run _after_ the `runMsgs` execution. They are also
	// defined as a chain, and have the same signature as antehandlers.
	//
	// In baseapp, postHandlers are run in the same store branch as `runMsgs`,
	// meaning that both `runMsgs` and `postHandler` state will be committed if
	// both are successful, and both will be reverted if any of the two fails.
	//
	// The SDK exposes a default postHandlers chain
	//
	// Please note that changing any of the anteHandler or postHandler chain is
	// likely to be a state-machine breaking change, which needs a coordinated
	// upgrade.
	app.setPostHandler()

	// At startup, after all modules have been registered, check that all proto
	// annotations are correct.
	protoFiles, err := proto.MergedRegistry()
	if err != nil {
		panic(err)
	}
	err = msgservice.ValidateProtoAnnotations(protoFiles)
	if err != nil {
		// Once we switch to using protoreflect-based antehandlers, we might
		// want to panic here instead of logging a warning.
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
	}

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			panic(fmt.Errorf("error loading last version: %w", err))
		}

		ctx := app.BaseApp.NewUncachedContext(true, tmproto.Header{})
		_ = ctx

		if err := app.WasmKeeper.InitializePinnedCodes(ctx); err != nil {
			panic(fmt.Sprintf("failed initialize pinned codes %s", err))
		}

		if err := wasmlckeeper.InitializePinnedCodes(ctx); err != nil {
			panic(fmt.Sprintf("wasmlckeeper failed initialize pinned codes %s", err))
		}
	}

	return app
}

func (app *ChainApp) FinalizeBlock(req *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error) {
	// when skipping sdk 47 for sdk 50, the upgrade handler is called too late in BaseApp
	// this is a hack to ensure that the migration is executed when needed and not panics
	app.once.Do(func() {
		ctx := app.NewUncachedContext(false, tmproto.Header{})
		if _, err := app.ConsensusParamsKeeper.Params(ctx, &consensusparamtypes.QueryParamsRequest{}); err != nil {
			// prevents panic: consensus key is nil: collections: not found: key 'no_key' of type github.com/cosmos/gogoproto/tendermint.types.ConsensusParams
			// sdk 47:
			// Migrate Tendermint consensus parameters from x/params module to a dedicated x/consensus module.
			// see https://github.com/cosmos/cosmos-sdk/blob/v0.47.0/simapp/upgrades.go#L66
			baseAppLegacySS := app.GetSubspace(baseapp.Paramspace)
			err := baseapp.MigrateParams(sdk.UnwrapSDKContext(ctx), baseAppLegacySS, app.ConsensusParamsKeeper.ParamsStore)
			if err != nil {
				panic(err)
			}
		}
	})

	return app.BaseApp.FinalizeBlock(req)
}

func (app *ChainApp) setAnteHandler(options chainante.HandlerOptions) {
	if err := options.Validate(); err != nil {
		panic(err)
	}

	app.SetAnteHandler(chainante.NewAnteHandler(options))
}

func (app *ChainApp) setPostHandler() {
	postHandler, err := posthandler.NewPostHandler(
		posthandler.HandlerOptions{},
	)
	if err != nil {
		panic(err)
	}

	app.SetPostHandler(postHandler)
}

// Name returns the name of the App
func (app *ChainApp) Name() string { return app.BaseApp.Name() }

// PreBlocker application updates every pre block
func (app *ChainApp) PreBlocker(ctx sdk.Context, _ *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
	return app.ModuleManager.PreBlock(ctx)
}

// BeginBlocker application updates every begin block
func (app *ChainApp) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	return app.ModuleManager.BeginBlock(ctx)
}

// EndBlocker application updates every end block
func (app *ChainApp) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	return app.ModuleManager.EndBlock(ctx)
}

func (a *ChainApp) Configurator() module.Configurator {
	return a.configurator
}

// InitChainer application update at chain initialization
func (app *ChainApp) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	var genesisState GenesisState
	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	err := app.UpgradeKeeper.SetModuleVersionMap(ctx, app.ModuleManager.GetVersionMap())
	if err != nil {
		panic(err)
	}
	response, err := app.ModuleManager.InitGenesis(ctx, app.appCodec, genesisState)
	return response, err
}

// LoadHeight loads a particular height
func (app *ChainApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// LegacyAmino returns legacy amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *ChainApp) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// AppCodec returns app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *ChainApp) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns ChainApp's InterfaceRegistry
func (app *ChainApp) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

// TxConfig returns ChainApp's TxConfig
func (app *ChainApp) TxConfig() client.TxConfig {
	return app.txConfig
}

// AutoCliOpts returns the autocli options for the app.
func (app *ChainApp) AutoCliOpts() autocli.AppOptions {
	modules := make(map[string]appmodule.AppModule, 0)
	for _, m := range app.ModuleManager.Modules {
		if moduleWithName, ok := m.(module.HasName); ok {
			moduleName := moduleWithName.Name()
			if appModule, ok := moduleWithName.(appmodule.AppModule); ok {
				modules[moduleName] = appModule
			}
		}
	}

	return autocli.AppOptions{
		Modules:               modules,
		ModuleOptions:         runtimeservices.ExtractAutoCLIOptions(app.ModuleManager.Modules),
		AddressCodec:          authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		ValidatorAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		ConsensusAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	}
}

// DefaultGenesis returns a default genesis from the registered AppModuleBasic's.
func (a *ChainApp) DefaultGenesis() map[string]json.RawMessage {
	genesis := a.BasicModuleManager.DefaultGenesis(a.appCodec)

	mintGenState := minttypes.DefaultGenesisState()
	mintGenState.Params.MintDenom = BaseDenom
	genesis[minttypes.ModuleName] = a.appCodec.MustMarshalJSON(mintGenState)

	return genesis
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *ChainApp) GetKey(storeKey string) *storetypes.KVStoreKey {
	return app.keys[storeKey]
}

// GetStoreKeys returns all the stored store keys.
func (app *ChainApp) GetStoreKeys() []storetypes.StoreKey {
	keys := make([]storetypes.StoreKey, 0, len(app.keys))
	for _, key := range app.keys {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Name() < keys[j].Name()
	})
	return keys
}

// GetTKey returns the TransientStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *ChainApp) GetTKey(storeKey string) *storetypes.TransientStoreKey {
	return app.tkeys[storeKey]
}

// GetMemKey returns the MemStoreKey for the provided mem key.
//
// NOTE: This is solely used for testing purposes.
func (app *ChainApp) GetMemKey(storeKey string) *storetypes.MemoryStoreKey {
	return app.memKeys[storeKey]
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (app *ChainApp) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// SimulationManager implements the SimulationApp interface
func (app *ChainApp) SimulationManager() *module.SimulationManager {
	return app.sm
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *ChainApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register new CometBFT queries routes from grpc-gateway.
	cmtservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register node gRPC service for grpc-gateway.
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register grpc-gateway routes for all modules.
	app.BasicModuleManager.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register swagger API from root so that other applications can override easily
	if err := server.RegisterSwaggerAPI(apiSvr.ClientCtx, apiSvr.Router, apiConfig.Swagger); err != nil {
		panic(err)
	}
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *ChainApp) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *ChainApp) RegisterTendermintService(clientCtx client.Context) {
	cmtApp := server.NewCometABCIWrapper(app)
	cmtservice.RegisterTendermintService(
		clientCtx,
		app.BaseApp.GRPCQueryRouter(),
		app.interfaceRegistry,
		cmtApp.Query,
	)
}

func (app *ChainApp) RegisterNodeService(clientCtx client.Context, cfg config.Config) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), cfg)
}

// GetMaccPerms returns a copy of the module account permissions
//
// NOTE: This is solely to be used for testing purposes.
func GetMaccPerms() map[string][]string {
	dupMaccPerms := make(map[string][]string)
	for k, v := range maccPerms {
		dupMaccPerms[k] = v
	}

	return dupMaccPerms
}

// BlockedAddresses returns all the app's blocked account addresses.
func BlockedAddresses() map[string]bool {
	blockedAddrs := make(map[string]bool)

	for acc := range GetMaccPerms() {
		blockedAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	// allow the following addresses to receive funds
	delete(blockedAddrs, authtypes.NewModuleAddress(govtypes.ModuleName).String())

	return blockedAddrs
}

func initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey storetypes.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	// required for testing finalized block migration
	paramsKeeper.Subspace(baseapp.Paramspace)

	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(minttypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(govtypes.ModuleName)
	paramsKeeper.Subspace(crisistypes.ModuleName)

	// register the IBC key tables for legacy param subspaces
	keyTable := ibcclienttypes.ParamKeyTable()
	keyTable.RegisterParamSet(&ibcconnectiontypes.Params{})
	paramsKeeper.Subspace(ibcexported.ModuleName).WithKeyTable(keyTable)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName).WithKeyTable(ibctransfertypes.ParamKeyTable())
	paramsKeeper.Subspace(icacontrollertypes.SubModuleName).WithKeyTable(icacontrollertypes.ParamKeyTable())
	paramsKeeper.Subspace(icahosttypes.SubModuleName).WithKeyTable(icahosttypes.ParamKeyTable())

	paramsKeeper.Subspace(wasmtypes.ModuleName)
	paramsKeeper.Subspace(packetforwardtypes.ModuleName)
	paramsKeeper.Subspace(ratelimittypes.ModuleName)

	return paramsKeeper
}
