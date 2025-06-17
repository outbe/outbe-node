package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/outbe/outbe-node/x/pool/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
)

func CmdQueryBlockEmission() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block-emission [block_number]",
		Short: "shows token emission amount every block",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			num, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				// Handle error if conversion fails
				fmt.Println("Error converting string to int64:", err)
				return
			}

			req := &types.QueryBlockEmissionRequest{
				BlockNumber: num,
			}

			res, err := queryClient.GetBlockEmission(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
