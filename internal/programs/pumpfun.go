package programs

import (
	"crypto/sha256"
	"pf-launcher/internal"
	"pf-launcher/internal/types"

	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/near/borsh-go"
)

func NewBuyIx(
	amount uint64,
	maxSolCost uint64,
	feeRecipient, mint,
	assocBondingCurve, assocUser,
	user, systemProgram, tokenProgram, creatorVault,
	eventAuthority solana.PublicKey,
) *solana.GenericInstruction {
	program := solana.MustPublicKeyFromBase58(internal.PUMP_FUN_PROGRAM)

	buyDiscriminator := sha256.Sum256([]byte("global:buy"))

	argsBin, _ := borsh.Serialize(types.BuyData{
		Amount:     amount,
		MaxSolCost: maxSolCost,
	})
	data := append(buyDiscriminator[:8], argsBin...)

	// Derive PDAs
	global, _, _ := DeriveGlobal(program)
	bondingCurve, _, _ := DeriveBondingCurve(mint, program)

	metas := solana.AccountMetaSlice{
		{PublicKey: global, IsWritable: false, IsSigner: false},
		{PublicKey: feeRecipient, IsWritable: true, IsSigner: false},
		{PublicKey: mint, IsWritable: false, IsSigner: false},
		{PublicKey: bondingCurve, IsWritable: true, IsSigner: false},
		{PublicKey: assocBondingCurve, IsWritable: true, IsSigner: false},
		{PublicKey: assocUser, IsWritable: true, IsSigner: false},
		{PublicKey: user, IsWritable: true, IsSigner: true},
		{PublicKey: systemProgram, IsWritable: false, IsSigner: false},
		{PublicKey: tokenProgram, IsWritable: false, IsSigner: false},
		{PublicKey: creatorVault, IsWritable: true, IsSigner: false},
		{PublicKey: eventAuthority, IsWritable: false, IsSigner: false},
		{PublicKey: program, IsWritable: false, IsSigner: false},
	}

	return &solana.GenericInstruction{
		AccountValues: metas,
		ProgID:        program,
		DataBytes:     data,
	}
}

func NewCreateIx(
	mint,
	user solana.PublicKey,
	createData types.CreateData,
) *solana.GenericInstruction {
	program := solana.MustPublicKeyFromBase58(internal.PUMP_FUN_PROGRAM)

	createDiscriminator := sha256.Sum256([]byte("global:create"))

	argsBin, _ := borsh.Serialize(createData)
	data := append(createDiscriminator[:8], argsBin...)

	mplTokenMetadata, _ := solana.PublicKeyFromBase58("metaqbxxUerdq28cj1RbAWkYQm3ybzjb6a8bt518x1s")
	eventAuthority, _ := solana.PublicKeyFromBase58("Ce6TQqeHC9p8KetsN6JsjHK7UTZk7nasjjnr7XxXp9F1")

	mintAuthority, _, _ := DeriveMintAuthority(program)
	bondingCurve, _, _ := DeriveBondingCurve(mint, program)
	global, _, _ := DeriveGlobal(program)
	associatedBondingCurve, _, _ := DeriveAssociatedBondingCurve(mint, bondingCurve)
	metadata, _, _ := DeriveMetadata(mint, mplTokenMetadata)

	metas := solana.AccountMetaSlice{
		{PublicKey: mint, IsWritable: true, IsSigner: true},
		{PublicKey: mintAuthority, IsWritable: false, IsSigner: false},
		{PublicKey: bondingCurve, IsWritable: false, IsSigner: false},
		{PublicKey: associatedBondingCurve, IsWritable: true, IsSigner: false},
		{PublicKey: global, IsWritable: true, IsSigner: false},
		{PublicKey: mplTokenMetadata, IsWritable: false, IsSigner: false},
		{PublicKey: metadata, IsWritable: true, IsSigner: false},
		{PublicKey: user, IsWritable: true, IsSigner: true},
		{PublicKey: solana.SystemProgramID, IsWritable: false, IsSigner: false},
		{PublicKey: solana.TokenProgramID, IsWritable: false, IsSigner: false},
		{PublicKey: associatedtokenaccount.ProgramID, IsWritable: false, IsSigner: false},
		{PublicKey: solana.SysVarRentPubkey, IsWritable: false, IsSigner: false},
		{PublicKey: eventAuthority, IsWritable: false, IsSigner: false},
		{PublicKey: program, IsWritable: false, IsSigner: false},
	}

	return &solana.GenericInstruction{
		AccountValues: metas,
		ProgID:        program,
		DataBytes:     data,
	}
}
