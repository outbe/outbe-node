package cli

import (
	"context"

	"github.com/outbe/outbe-node/x/gemmint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
)

func CmdListWhitelist() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "whitelist",
		Short: "shows whitelist details",
		// Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			// argId := args[0]

			params := &types.QueryWhitelistRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.Whitelist(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
