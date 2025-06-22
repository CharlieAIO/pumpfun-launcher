package programs

import (
	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
)

// DeriveMintAuthority derives the mint_authority PDA
// Seeds: ["mint-authority"]
func DeriveMintAuthority(programID solana.PublicKey) (solana.PublicKey, uint8, error) {
	return solana.FindProgramAddress(
		[][]byte{
			[]byte("mint-authority"),
		},
		programID,
	)
}

// DeriveBondingCurve derives the bonding_curve PDA
// Seeds: ["bonding-curve", mint]
func DeriveBondingCurve(mint, programID solana.PublicKey) (solana.PublicKey, uint8, error) {
	return solana.FindProgramAddress(
		[][]byte{
			[]byte("bonding-curve"),
			mint.Bytes(),
		},
		programID,
	)
}

// DeriveGlobal derives the global PDA
// Seeds: ["global"]
func DeriveGlobal(programID solana.PublicKey) (solana.PublicKey, uint8, error) {
	return solana.FindProgramAddress(
		[][]byte{
			[]byte("global"),
		},
		programID,
	)
}

// DeriveAssociatedBondingCurve derives the associated_bonding_curve PDA
// Seeds: [bondingCurve, tokenProgramID, mint]
func DeriveAssociatedBondingCurve(mint, bondingCurve solana.PublicKey) (solana.PublicKey, uint8, error) {
	return solana.FindProgramAddress(
		[][]byte{
			bondingCurve.Bytes(),
			solana.TokenProgramID.Bytes(),
			mint.Bytes(),
		},
		solana.PublicKey(associatedtokenaccount.ProgramID.Bytes()),
	)
}

// DeriveMetadata derives the metadata PDA
// Seeds: ["metadata", mplTokenMetadata, mint]
func DeriveMetadata(mint, mplTokenMetadata solana.PublicKey) (solana.PublicKey, uint8, error) {
	return solana.FindProgramAddress(
		[][]byte{
			[]byte("metadata"),
			mplTokenMetadata.Bytes(),
			mint.Bytes(),
		},
		solana.PublicKey(mplTokenMetadata.Bytes()),
	)

}

// DeriveAssociatedTokenAccount derives the associated token account
// Seeds: [wallet, tokenProgramID, mint]
func DeriveAssociatedTokenAccount(wallet, mint solana.PublicKey) (solana.PublicKey, uint8, error) {
	return solana.FindProgramAddress(
		[][]byte{
			wallet.Bytes(),
			solana.TokenProgramID.Bytes(),
			mint.Bytes(),
		},
		associatedtokenaccount.ProgramID,
	)
}

func DeriveCreatorVault(creator, programID solana.PublicKey) (solana.PublicKey, uint8, error) {
	return solana.FindProgramAddress(
		[][]byte{
			[]byte("creator-vault"),
			creator.Bytes(),
		},
		programID,
	)
}
