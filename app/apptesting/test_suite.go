package apptesting

import (
	"github.com/outbe/outbe-node/app"
	"github.com/outbe/outbe-node/app/params"
	"github.com/outbe/outbe-node/x/pool/types"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storemetrics "cosmossdk.io/store/metrics"
	"cosmossdk.io/store/rootmulti"
	storetypes "cosmossdk.io/store/types"
	"github.com/cometbft/cometbft/crypto/ed25519"
	tenderminttypes "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/stretchr/testify/suite"
)

type KeeperTestSuite struct {
	suite.Suite
	Ctx         sdk.Context
	App         *app.ChainApp
	QueryClient types.QueryClient
	QueryHelper *baseapp.QueryServiceTestHelper
	TestAccs    []sdk.AccAddress
}

func (s *KeeperTestSuite) Setup() {
	s.App = app.Setup(s.T())
	s.Ctx = s.App.BaseApp.NewContext(false)

	s.QueryHelper = &baseapp.QueryServiceTestHelper{
		GRPCQueryRouter: s.App.GRPCQueryRouter(),
		Ctx:             s.Ctx,
	}
	s.TestAccs = s.CreateRandomAccounts(3)
}

// CreateRandomAccounts is a function return a list of randomly generated AccAddresses
func (s *KeeperTestSuite) CreateRandomAccounts(numAccts int) []sdk.AccAddress {
	testAddrs := make([]sdk.AccAddress, numAccts)
	for i := 0; i < numAccts; i++ {
		pk := ed25519.GenPrivKey().PubKey()
		testAddrs[i] = sdk.AccAddress(pk.Address())
	}
	return testAddrs
}

func (s *KeeperTestSuite) FundAccount(acc sdk.AccAddress, amounts math.Int) {

	ctx := s.App.NewContext(true)
	coins := sdk.Coins{sdk.NewCoin(params.BondDenom, amounts)}
	err := simtestutil.FundModuleAccount(ctx, s.App.BankKeeper, "pool", coins)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) FundModuleAcc(moduleName string, amounts sdk.Coins) {
	err := simtestutil.FundModuleAccount(s.Ctx, s.App.BankKeeper, moduleName, amounts)
	s.Require().NoError(err)
}

// func (s *KeeperTestSuite) MintCoins(coins sdk.Coins) {
// 	err := s.App.BankKeeper.MintCoins(s.Ctx, ?.ModuleName, coins)
// 	s.Require().NoError(err)
// }

func (s *KeeperTestSuite) CreateTestContextWithMultiStore() (sdk.Context, storetypes.CommitMultiStore) {
	db := dbm.NewMemDB()
	logger := log.NewNopLogger()

	ms := rootmulti.NewStore(db, logger, storemetrics.NewNoOpMetrics())

	return sdk.NewContext(ms, tenderminttypes.Header{}, false, logger), ms
}
