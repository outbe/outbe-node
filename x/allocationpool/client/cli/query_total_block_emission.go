package cli

import (
	"context"
	"strconv"

	"github.com/outbe/outbe-node/x/allocationpool/types"

	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	errortypes "github.com/outbe/outbe-node/errors"
	"github.com/spf13/cobra"
)

func CmdQueryBlockEmission() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "total-block-emission [block_number]",
		Short: "Query token emission amount for a specific block",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return sdkerrors.Wrap(err, "failed to get client context")
			}

			queryClient := types.NewQueryClient(clientCtx)

			blockNumber, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return sdkerrors.Wrapf(errortypes.ErrInvalidRequest, "failed to parse block number: %v", err)
			}

			req := &types.QueryTotalBlockEmissionRequest{
				BlockNumber: blockNumber,
			}

			res, err := queryClient.GetTotalBlockEmission(context.Background(), req)
			if err != nil {
				return sdkerrors.Wrapf(err, "failed to query block emission for block %d", blockNumber)
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
