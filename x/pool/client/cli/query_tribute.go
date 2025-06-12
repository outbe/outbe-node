package cli

import (
	"context"

	"github.com/outbe/outbe-node/x/pool/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
)

func CmdQueryTribute() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-tribute",
		Short: "list all tribute",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryTributeRequest{}

			res, err := queryClient.GetTribute(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
