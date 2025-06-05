package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/outbe/outbe-node/x/gemmint/types"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdAddContractToWhitelist() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-contract-to-whitelist [contract_address]",
		Short: "Register a contract which is eligible to mint native token",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			fmt.Println("000000000000000000000000000000000")
			fmt.Println("000000000000000000000000000000000", args[0])

			// Create MsgAddContractToWhitelist
			msg := &types.MsgAddContractToWhitelist{
				Creator:         clientCtx.GetFromAddress().String(),
				ContractAddress: args[0],
			}

			// Broadcast the transaction
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	// Add transaction flags
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
