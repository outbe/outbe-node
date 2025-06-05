package keeper_test

import (
	"log"
	"time"

	"github.com/outbe/outbe-node/x/gemmint/types"
	_ "github.com/stretchr/testify/suite"
)

func (suite *KeeperTestHelper) TestWhitelist() {

	suite.SetupTest()
	ctx := suite.Ctx

	whitelist := types.Whitelist{
		Creator: "gem1hj5fveer5cjtn4wd6wstzugjfdxzl0xp8h9ghz",
		EligibleContracts: []*types.EligibleContract{
			{
				ContractAddress: "soar14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sg0qwca",
				TargetMint:      0,
				TotalMinted:     20,
				Created:         time.Now().UTC().Format(time.RFC3339),
				Enabled:         true,
			},
		},
		TotalMinted: 1000,
		Created:     time.Now().UTC().Format(time.RFC3339),
	}
	suite.App.MintKeeper.SetWhitelist(ctx, whitelist)
	whitelistDetails := suite.App.MintKeeper.GetWhitelist(ctx)
	log.Println("Whitelist Details", whitelistDetails[0])
}
