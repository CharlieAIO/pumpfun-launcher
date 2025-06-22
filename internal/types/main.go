package types

import (
	"fmt"
	"math/big"

	"github.com/gagliardetto/solana-go"
)

type BuyData struct {
	Amount     uint64
	MaxSolCost uint64
}

type CreateData struct {
	Name    string           `bson:"name"`
	Symbol  string           `bson:"symbol"`
	Uri     string           `bson:"uri"`
	Creator solana.PublicKey `bson:"creator"`
}

type Metadata struct {
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Description string `json:"description"`
	Image       string `json:"image"`
	ShowName    bool   `json:"showName"`
	CreatedOn   string `json:"createdOn"`
	Twitter     string `json:"twitter"`
	Telegram    string `json:"telegram"`
	Website     string `json:"website"`
}

type GlobalAccount struct {
	Discriminator               uint64           `borsh:"discriminator"`
	Initialized                 bool             `borsh:"initialized"`
	Authority                   solana.PublicKey `borsh:"authority"`
	FeeRecipient                solana.PublicKey `borsh:"fee_recipient"`
	InitialVirtualTokenReserves uint64           `borsh:"initial_virtual_token_reserves"`
	InitialVirtualSolReserves   uint64           `borsh:"initial_virtual_sol_reserves"`
	InitialRealTokenReserves    uint64           `borsh:"initial_real_token_reserves"`
	TokenTotalSupply            uint64           `borsh:"token_total_supply"`
	FeeBasisPoints              uint64           `borsh:"fee_basis_points"`
}

func (g *GlobalAccount) GetInitialBuyPrice(solAmount uint64) (uint64, error) {
	if solAmount <= 0 {
		return 0, nil
	}

	vSol := new(big.Int).SetUint64(g.InitialVirtualSolReserves)
	vToken := new(big.Int).SetUint64(g.InitialVirtualTokenReserves)

	// Add 5% buffer to solAmount for slippage
	amount := new(big.Int).SetUint64(solAmount)
	buffer := new(big.Int).Div(amount, big.NewInt(20))
	amount = new(big.Int).Add(amount, buffer)

	// Calculate k = x * y
	k := new(big.Int).Mul(vSol, vToken)

	// Calculate new sol reserves: i = x + amount
	newSolReserves := new(big.Int).Add(vSol, amount)

	// Calculate r = k/i (rounded up)
	r := new(big.Int).Div(k, newSolReserves)
	r.Add(r, big.NewInt(1)) // Add 1 to handle division rounding

	// Calculate s = vToken - r
	s := new(big.Int).Sub(vToken, r)

	// Check if s is negative
	if s.Sign() < 0 {
		return 0, fmt.Errorf("negative token amount calculated")
	}

	// Convert back to uint64, checking for overflow
	if !s.IsUint64() {
		return 0, fmt.Errorf("token amount overflow")
	}

	result := s.Uint64()
	if result < g.InitialRealTokenReserves {
		return result, nil
	}
	return g.InitialRealTokenReserves, nil
}
