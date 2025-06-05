package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/outbe/outbe-node/x/gemmint/types"
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
	//cmd.AddCommand(CmdQueryMinters())
	cmd.AddCommand(CmdListWhitelist())
	cmd.AddCommand(CmdMinted())
	// this line is used by starport scaffolding # 1

	return cmd
}

// func CmdQueryMinters() *cobra.Command {
// 	cmd := &cobra.Command{
// 		Use:   "minters",
// 		Short: "fetch all minters",
// 		Args:  cobra.NoArgs,
// 		RunE: func(cmd *cobra.Command, _ []string) error {
// 			clientCtx, err := client.GetClientQueryContext(cmd)
// 			if err != nil {
// 				return err
// 			}

// 			queryClient := types.NewQueryClient(clientCtx)

// 			pageReq, err := client.ReadPageRequest(cmd.Flags())
// 			if err != nil {
// 				return err
// 			}

// 			req := &types.QueryMinterAllRequest{
// 				Pagination: pageReq,
// 			}

// 			res, err := queryClient.Minters(cmd.Context(), req)
// 			if err != nil {
// 				return err
// 			}

// 			return clientCtx.PrintProto(res)
// 		},
// 	}

// 	flags.AddQueryFlagsToCmd(cmd)

// 	return cmd
// }
