package stellar

import "github.com/shopspring/decimal"

type Balance struct {
	AssetType   string // "native", "credit_alphanum4", "credit_alphanum12"
	AssetCode   string // "XLM", "USDC", etc
	AssetIssuer string // Issuer public key (empty for native)
	Amount      decimal.Decimal
	Limit       *decimal.Decimal // For trustlines
}
