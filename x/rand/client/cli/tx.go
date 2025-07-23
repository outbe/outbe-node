package cli

import (
	"crypto/sha256"
	"encoding/hex"
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
		GenerateCommitHashCmd(),
	)

	return txCmd
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

// ComputeHash (must match the Reveal handler's implementation)
func ComputeHash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

func NewGenerateCommitHashCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate-commit-hash [reveal-value] --output [format]",
		Short: "Generate a commitment hash from a reveal value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			outputFormat, err := cmd.Flags().GetString("output")
			if err != nil {
				return err
			}

			// Parse the reveal value
			bytes, err := ParseByteSlice(args[0])
			if err != nil {
				return fmt.Errorf("failed to parse reveal value: %w", err)
			}

			// Compute the hash (must match the Reveal handler's ComputeHash)
			hash := ComputeHash(bytes)

			// Output based on format
			switch outputFormat {
			case "hex":
				fmt.Println(hex.EncodeToString(hash))
			case "byte":
				byteStr := make([]string, len(hash))
				for i, b := range hash {
					byteStr[i] = fmt.Sprintf("%d", b)
				}
				fmt.Printf("[%s]\n", strings.Join(byteStr, " "))
			default:
				return fmt.Errorf("unsupported output format: %s", outputFormat)
			}

			return nil
		},
	}

	cmd.Flags().String("output", "hex", "Output format (hex or byte)")
	return cmd
}

// ParseByteSlice (from your existing code)
func ParseByteSlice(input string) ([]byte, error) {
	clean := strings.Trim(input, "[]")
	parts := strings.Fields(clean)
	result := make([]byte, 0, len(parts))
	for _, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid byte value %q: %w", part, err)
		}
		if num < 0 || num > 255 {
			return nil, fmt.Errorf("byte value out of range (0-255): %d", num)
		}
		result = append(result, byte(num))
	}
	return result, nil
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

func GenerateCommitHashCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate-commit-hash [reveal-value] --output [format]",
		Short: "Generate a commitment hash from a reveal value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			outputFormat, err := cmd.Flags().GetString("output")
			if err != nil {
				return err
			}

			// Parse the reveal value
			bytes, err := ParseByteSlice(args[0])
			if err != nil {
				return fmt.Errorf("failed to parse reveal value: %w", err)
			}

			// Compute the hash (must match the ComputeHash function used in Reveal)
			hash := ComputeHash(bytes)

			// Output based on format
			switch outputFormat {
			case "hex":
				fmt.Println(hex.EncodeToString(hash))
			case "byte":
				fmt.Printf("[%s]\n", strings.Join(strings.Fields(fmt.Sprintf("%d", hash)), " "))
			default:
				return fmt.Errorf("unsupported output format: %s", outputFormat)
			}

			return nil
		},
	}

	cmd.Flags().String("output", "hex", "Output format (hex or byte)")
	return cmd
}
