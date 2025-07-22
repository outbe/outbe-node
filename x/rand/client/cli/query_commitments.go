package cli

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	errortypes "github.com/outbe/outbe-node/errors"
	"github.com/outbe/outbe-node/x/rand/types"
	"github.com/spf13/cobra"
)

func CmdQueryCommitment() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commitment [period] [validator]",
		Short: "Query commitments for a given period and validator",
		Args:  cobra.ExactArgs(2), // Expect exactly two arguments: period and validator
		RunE: func(cmd *cobra.Command, args []string) error {

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			period := args[0]
			validator := args[1]

			// Create the query request
			req := &types.QueryCommitmentRequest{
				Period:    period,
				Validator: validator,
			}

			// Query the gRPC endpoint
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Commitment(cmd.Context(), req)
			if err != nil {
				return sdkerrors.Wrap(errortypes.ErrNoCommitment, "failed to query commitments")
			}

			// Output the response
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdQueryCommitments() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commitments",
		Short: "shows the commitments state of the module",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Commitments(context.Background(), &types.QueryCommitmentsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
