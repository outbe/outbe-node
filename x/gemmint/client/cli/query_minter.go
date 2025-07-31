package cli

import (
	"context"

	"github.com/outbe/outbe-node/x/gemmint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
)

func CmdMinter() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "minter",
		Short: "shows minter details",
		// Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)

			// pageReq, err := client.ReadPageRequest(cmd.Flags())
			// if err != nil {
			// 	return err
			// }

			queryClient := types.NewQueryClient(clientCtx)

			// argId := args[0]

			params := &types.QueryMinterRequest{}

			res, err := queryClient.Minters(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
