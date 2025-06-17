package keeper_test

import (
	"testing"

	"github.com/outbe/outbe-node/x/pool/types"

	"github.com/outbe/outbe-node/app/apptesting"

	"github.com/stretchr/testify/suite"
)

type KeeperTestHelper struct {
	apptesting.KeeperTestSuite
	queryClient types.QueryClient
}

func (suite *KeeperTestHelper) SetupTest() {
	suite.Setup()
	suite.queryClient = types.NewQueryClient(suite.QueryHelper)
}
func TestKeeperTestHelper(t *testing.T) {
	suite.Run(t, new(KeeperTestHelper))
}
