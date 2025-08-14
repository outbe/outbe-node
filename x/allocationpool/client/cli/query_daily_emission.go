package cli

import (
	"context"

	"github.com/outbe/outbe-node/x/allocationpool/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
)

func CmdQueryDailyEmission() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "daily-emission",
		Short: "shows tokens daily emission amount",
		//Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryDailyEmissionRequest{}

			res, err := queryClient.GetDailyEmission(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
