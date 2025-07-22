package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/outbe/outbe-node/x/rand/types"
	"github.com/spf13/cobra"
)

func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the rand module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	queryCmd.AddCommand(
		CmdQueryParams(),
		CmdQueryPeriod(),
		CmdQueryReveals(),
		CmdQueryCommitment(),
		CmdQueryCommitments(),
		CmdQueryCurrentRandomness(),
	)

	return queryCmd
}
