package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/outbe/outbe-node/x/rand/types"
	"github.com/spf13/cobra"
)

func CmdQueryPenalty() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "penalties",
		Short: "shows the penalized validators",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Penalties(context.Background(), &types.QueryPenaltiesRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
