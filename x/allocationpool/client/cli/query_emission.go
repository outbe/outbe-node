package cli

import (
	"context"
	"strconv"

	"github.com/outbe/outbe-node/x/allocationpool/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	errortypes "github.com/outbe/outbe-node/errors"
	"github.com/spf13/cobra"

	sdkerrors "cosmossdk.io/errors"
)

func CmdQueryEmissions() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-block-emission",
		Short: "shows all emission per block amount",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryAllEmissionRequest{}

			res, err := queryClient.EmissionAll(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryEmissionEntity() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block-emission [block_number]",
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

			req := &types.QueryEmissionEntityRequest{
				BlockNumber: blockNumber,
			}

			res, err := queryClient.GetEmissionEntity(context.Background(), req)
			if err != nil {
				return sdkerrors.Wrapf(err, "failed to query block emission for block %d", blockNumber)
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
