package cli

import (
	"context"

	"github.com/outbe/outbe-node/x/pool/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
)

func CmdQueryEmission() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "total-emission",
		Short: "shows tokens total emission amount",
		//Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryEmissionRequest{}

			res, err := queryClient.GetEmission(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
