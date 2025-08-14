package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/outbe/outbe-node/x/allocationpool/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	// Group minting queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryParams())
	cmd.AddCommand(CmdQueryBlockEmission())
	cmd.AddCommand(CmdQueryLimit())
	cmd.AddCommand(CmdQueryEmissions())
	cmd.AddCommand(CmdQueryTributes())
	cmd.AddCommand(CmdQueryEmissionEntity())
	cmd.AddCommand(CmdQueryAllCRAs())
	cmd.AddCommand(CmdQueryAllWallets())
	// this line is used by starport scaffolding # 1

	return cmd
}
