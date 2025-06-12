package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/outbe/outbe-node/x/pool/types"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdMintTribute() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mint_tribute [contract_address] [receipt_address] [amount]",
		Short: "Mint & Distribute tributes to smart contracts.",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			coin, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return err
			}

			// Create MsgDistributeAuctionReward
			msg := &types.MsgMintTribute{
				Creator:         clientCtx.GetFromAddress().String(),
				ContractAddress: args[0],
				MintAmount:      coin,
				ReceiptAddress:  args[1],
			}

			// Broadcast the transaction
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	// Add transaction flags
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
