package cli

import (
	"fmt"
	"strconv"
	"strings"

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

func ParseByteSlice(input string) ([]byte, error) {
	// Trim brackets
	clean := strings.Trim(input, "[]")

	// Split by whitespace
	parts := strings.Fields(clean)

	// Convert to []byte
	result := make([]byte, 0, len(parts))
	for _, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid byte value %q: %w", part, err)
		}
		result = append(result, byte(num))
	}

	return result, nil
}

func NewCommitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit [validator] [commitment-hash] [deposit]",
		Short: "Submit a commitment",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {

			bytes, err := ParseByteSlice(args[1])
			if err != nil {
				panic(err)
			}

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

			msg := types.NewMsgCommit(clientCtx.GetFromAddress().String(), validator, bytes, deposit)
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

			bytes, err := ParseByteSlice(args[2])
			if err != nil {
				panic(err)
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			period, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgReveal(clientCtx.GetFromAddress().String(), period, args[1], bytes)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
