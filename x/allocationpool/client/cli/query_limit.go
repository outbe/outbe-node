package cli

import (
	"context"

	"github.com/outbe/outbe-node/x/allocationpool/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
)

func CmdQueryLimit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "limit",
		Short: "fetch current limit",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryLimitRequest{}

			res, err := queryClient.GetLimit(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
