package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/outbe/outbe-node/x/rand/types"
	"github.com/spf13/cobra"
)

func CmdQueryReveals() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reveals",
		Short: "shows the reveals state of the module",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Reveals(context.Background(), &types.QueryRevealsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
