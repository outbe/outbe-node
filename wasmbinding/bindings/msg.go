package bindings

type QcoreMsg struct {
	MsgMintTribute *MsgMintTribute `json:"msg_mint_tribute,omitempty"`
}

type MsgMintTribute struct {
	Creator         string `json:"creator"`
	ContractAddress string `json:"contract_address"`
	MintAmount      string `json:"mint_amount"`
	ReceiptAddress  string `json:"receipt_address"`
}
