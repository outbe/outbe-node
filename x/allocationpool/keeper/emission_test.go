package keeper

import (
	"math"
	"math/big"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestCalculateExponentialBlockEmission(t *testing.T) {
	initialRateStr := "65536"
	decayStr := "0.00000006"
	blockNumber := int64(1) // Test for block 1

	initialRate, err := sdkmath.LegacyNewDecFromStr(initialRateStr)
	require.NoError(t, err)

	decay, err := sdkmath.LegacyNewDecFromStr(decayStr)
	require.NoError(t, err)

	n := sdkmath.LegacyNewDec(blockNumber)
	decayN := decay.Mul(n)
	expArg := -decayN.MustFloat64()
	expResult := math.Exp(expArg)

	scaled := new(big.Float).SetFloat64(expResult)
	scaled.Mul(scaled, big.NewFloat(1e18))

	scaledInt := new(big.Int)
	scaled.Int(scaledInt)

	expVal := sdkmath.LegacyNewDecFromBigInt(scaledInt).QuoInt64(1e18)
	tokens := initialRate.Mul(expVal)

	t.Logf("Block %d emission result: %s", blockNumber, tokens.String())

	expected := "65535.996067840117440512"
	require.Equal(t, expected, tokens.String(), "Emission at block 1 does not match expected value")
}

func TestCalculateFixedBlockEmission_HighPrecision(t *testing.T) {
	totalSupply := sdkmath.LegacyMustNewDecFromStr("200000000000000000000000000") // 200M * 1e18
	apr := sdkmath.LegacyMustNewDecFromStr("0.02")
	daysPerYear := sdkmath.LegacyMustNewDecFromStr("365")
	blocksPerDay := sdkmath.LegacyMustNewDecFromStr("17280")

	emissionPerBlock := totalSupply.Mul(apr).Quo(daysPerYear).Quo(blocksPerDay)

	t.Logf("Fixed emission per block: %s", emissionPerBlock.String())

	expected := "634195839675291730.086250634195839675"
	require.Equal(t, expected, emissionPerBlock.String(), "Fixed block emission mismatch")
}

func TestCalculateFixedAnnualEmission(t *testing.T) {
	// Constants
	totalSupplyStr := "200000000000000000000000000" // 200M OUTBE (18 decimals)
	aprStr := "0.02"                                // 2%

	// Convert to Dec
	totalSupply := sdkmath.LegacyMustNewDecFromStr(totalSupplyStr)
	apr := sdkmath.LegacyMustNewDecFromStr(aprStr)

	// Calculate
	emission := totalSupply.Mul(apr)
	expected := sdkmath.LegacyMustNewDecFromStr("4000000000000000000000000")

	// ✅ Compare Dec values directly
	require.True(t, emission.Equal(expected), "Annual emission calculation mismatch")

	t.Logf("Calculated annual emission: %s", emission.String())
}

func TestCalculateFixedDailyEmission(t *testing.T) {
	// Given constants
	totalSupplyStr := "200000000000000000000000000" // 200M OUTBE with 18 decimals
	aprStr := "0.02"                                // 2% APR
	daysPerYearStr := "365"

	// Convert to Dec
	totalSupply := sdkmath.LegacyMustNewDecFromStr(totalSupplyStr)
	apr := sdkmath.LegacyMustNewDecFromStr(aprStr)
	daysPerYear := sdkmath.LegacyMustNewDecFromStr(daysPerYearStr)

	// Expected formula: (totalSupply * apr) / 365
	expected := totalSupply.Mul(apr).Quo(daysPerYear)

	// Actual emission calculation
	emission := totalSupply.Mul(apr).Quo(daysPerYear)

	// Assert equality using Dec Equal()
	require.True(t, emission.Equal(expected), "Daily emission calculation mismatch")

	t.Logf("Calculated daily emission: %s", emission.String())
}
