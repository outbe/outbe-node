package keeper_test

import (
	"fmt"
	"strconv"

	sdkmath "cosmossdk.io/math"
	"github.com/outbe/outbe-node/x/pool/constants"
)

func (suite *KeeperTestHelper) TestCalculateExponentialBlockEmission() {
	suite.SetupTest()

	testCases := []struct {
		blockHeight   int64
		expectedToken float64
	}{
		{blockHeight: 1, expectedToken: 18475.315250812559},
		{blockHeight: 2, expectedToken: 18475.314327046817},
		{blockHeight: 3, expectedToken: 18475.313403281125},
		{blockHeight: 4, expectedToken: 18475.312479515476},
		{blockHeight: 5, expectedToken: 18475.31155574988},
		{blockHeight: 6, expectedToken: 18475.310631984325},
		{blockHeight: 7, expectedToken: 18475.309708218814},
		{blockHeight: 8, expectedToken: 18475.30878445335},
		{blockHeight: 9, expectedToken: 18475.307860687935},
		{blockHeight: 10, expectedToken: 18475.306936922567}, // Adjust based on actual expected values
		// ...
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("BlockHeight=%d", tc.blockHeight), func() {
			tokenStr, err := suite.App.PoolKeeper.CalculateExponentialBlockEmission(tc.blockHeight)

			// Assert no error from the function
			suite.Require().NoError(err, "CalculateExponentialTokens should not return an error")

			// Ensure token string is not empty
			suite.Require().NotEmpty(tokenStr, "Token string should not be empty")

			// Parse token string as float64
			token, err := strconv.ParseFloat(tokenStr, 64)
			suite.Require().NoError(err, "Failed to convert token string to float64")

			// Assert the token is positive
			suite.Require().Greater(token, 0.0, "Token amount should be positive")

			// Optional: If tokens should be whole numbers in practice, check if the decimal part is negligible
			// This depends on the expected behavior of CalculateExponentialTokens

			suite.Require().Equal(tc.expectedToken, token, "Unexpected token amount for block %d", tc.blockHeight)

			// Log for debugging
			suite.T().Logf("Exponential Token emission for block %d: %s (%f)", tc.blockHeight, tokenStr, token)
		})
	}
}
func (suite *KeeperTestHelper) TestCalculateFixedBlockEmission() {

	suite.SetupTest()

	testCases := []struct {
		totalSupplyStr   string
		expectedEmission float64 // Based on manual calculation
	}{
		{"1000000000000000000000000", 315.6597}, // 1 million tokens
		{"500000000000000000000000", 157.8298},  // 0.5 million
		{"10000000000000000000000", 3.1566},     // 10k tokens
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("TotalSupply=%s", tc.totalSupplyStr), func() {

			tokenStr, err := calculateFixedBlockEmission(tc.totalSupplyStr)

			suite.Require().NoError(err, "CalculateFixedBlockEmission should not return an error")
			suite.Require().NotEmpty(tokenStr, "Token string should not be empty")

		})
	}
}

func calculateFixedBlockEmission(totalSupply string) (string, error) {

	// Fixed emission: (totalSupply * 0.02) / 365 / 17280
	apr, err := sdkmath.LegacyNewDecFromStr(constants.APR)
	if err != nil {
		return "", err
	}

	daysPerYear, err := sdkmath.LegacyNewDecFromStr(constants.DaysPerYear)
	if err != nil {
		return "", err
	}

	blocksPerDay, err := sdkmath.LegacyNewDecFromStr(constants.BlocksPerDay)
	if err != nil {
		return "", err
	}

	if totalSupply <= "0" {
		return "", fmt.Errorf("total supply must be greater than zero")
	}

	decTotalSupply := sdkmath.LegacyMustNewDecFromStr(totalSupply)
	emissionPerBlock := decTotalSupply.Mul(apr).Quo(daysPerYear).Quo(blocksPerDay)

	fmt.Printf("emissionPerBlock is %s for total supply = %s\n", emissionPerBlock.String(), totalSupply)

	return emissionPerBlock.String(), nil
}

func (suite *KeeperTestHelper) TestGetEmission() {

	suite.SetupTest()

	emission, found := suite.App.PoolKeeper.GetTotalEmission(suite.Ctx)

	fmt.Println("emission ------->", emission)

	suite.Require().Equal(found, true)
	suite.Require().NotEmpty(emission)
}
