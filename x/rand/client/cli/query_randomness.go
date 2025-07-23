package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/outbe/outbe-node/x/rand/types"
	"github.com/spf13/cobra"
)

func CmdQueryCurrentRandomness() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "current-randomness",
		Short: "Query the current randomness",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.CurrentRandomness(cmd.Context(), &types.QueryCurrentRandomnessRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
