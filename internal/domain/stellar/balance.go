package stellar

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type Balance struct {
	AssetType   string // "native", "credit_alphanum4", "credit_alphanum12"
	AssetCode   string // "XLM", "USDC", etc
	AssetIssuer string // Issuer public key (empty for native)
	Amount      decimal.Decimal
	Limit       *decimal.Decimal // For trustlines
}

func (b *Balance) Validate() error {
	if b.Amount.IsNegative() {
		return fmt.Errorf("amount cannot be negative")
	}

	if b.AssetType == "" {
		return fmt.Errorf("asset type is required")
	}

	validTypes := map[string]bool{
		"native":            true,
		"credit_alphanum4":  true,
		"credit_alphanum12": true,
	}

	if !validTypes[b.AssetType] {
		return fmt.Errorf("invalid asset type: %s", b.AssetType)
	}

	if b.AssetType != "native" {
		if b.AssetCode == "" {
			return fmt.Errorf("asset code is required for non-native assets")
		}
		if b.AssetIssuer == "" {
			return fmt.Errorf("asset issuer is required for non-native assets")
		}
		if !strings.HasPrefix(b.AssetIssuer, "G") {
			return fmt.Errorf("invalid issuer public key format")
		}
	}

	return nil
}
