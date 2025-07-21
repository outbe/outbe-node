package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/outbe/outbe-node/x/rand/types"
	"github.com/spf13/cobra"
)

func GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Transaction commands for the rand module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewCommitCmd(),
		NewRevealCmd(),
	)

	return txCmd
}

func NewCommitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit [validator] [commitment-hash] [deposit]",
		Short: "Submit a commitment",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			validator, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			deposit, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return err
			}

			msg := types.NewMsgCommit(clientCtx.GetFromAddress().String(), validator, args[1], deposit)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func NewRevealCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reveal [period] [validator] [reveal-value]",
		Short: "Submit a reveal",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {

			fmt.Println("here---------------------")

			fmt.Println("NewRevealCmd ----------args------- >", args[0], args[1], args[2])

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgReveal(clientCtx.GetFromAddress().String(), args[0], args[1], args[2])
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
