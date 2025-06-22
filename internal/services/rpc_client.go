package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/near/borsh-go"

	"pf-launcher/internal"
	"pf-launcher/internal/programs"
	"pf-launcher/internal/types"

	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
	"github.com/joho/godotenv"
)

type RPCClient struct {
	rpcClient *rpc.Client
	user      *solana.Wallet
	mint      *solana.Wallet
}

func NewRPCClient(privateKey string) (*RPCClient, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	rpcURL := os.Getenv("RPC")
	if rpcURL == "" {
		return nil, fmt.Errorf("RPC URL not found in .env file")
	}

	rpcClient := rpc.New(rpcURL)

	user, err := solana.WalletFromPrivateKeyBase58(privateKey)
	if err != nil {
		return nil, fmt.Errorf("error creating wallet: %v", err)
	}

	return &RPCClient{
		rpcClient: rpcClient,
		user:      user,
	}, nil
}

func (c *RPCClient) LaunchToken(metadata types.Metadata, metadataUri string, solAmount uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	createIx, err := c.AddCreateInstruction(metadata, metadataUri)
	if err != nil {
		return fmt.Errorf("failed to add create instruction: %w", err)
	}

	createAssocIx := associatedtokenaccount.NewCreateInstruction(
		c.user.PublicKey(),
		c.user.PublicKey(),
		c.mint.PublicKey(),
	).Build()

	buyIx, err := c.AddBuyInstruction(c.mint.PublicKey(), solAmount)
	if err != nil {
		return fmt.Errorf("failed to add buy instruction: %w", err)
	}

	var bh *rpc.GetLatestBlockhashResult
	for i := 0; i < 3; i++ {
		bh, err = c.rpcClient.GetLatestBlockhash(ctx, rpc.CommitmentProcessed)
		if err == nil {
			break
		}
		log.Printf("Attempt %d: Failed to get blockhash: %v", i+1, err)
		time.Sleep(time.Second * time.Duration(i+1))
	}
	if err != nil {
		return fmt.Errorf("failed to get blockhash after retries: %w", err)
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{createIx, createAssocIx, buyIx},
		bh.Value.Blockhash,
		solana.TransactionPayer(c.user.PublicKey()),
	)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	log.Printf("mint: %+v", c.mint.PublicKey())
	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(c.user.PublicKey()) {
			return &c.user.PrivateKey
		}
		if key.Equals(c.mint.PublicKey()) {
			return &c.mint.PrivateKey
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Retry sending transaction
	var sig solana.Signature
	for i := 0; i < 3; i++ {
		sig, err = c.rpcClient.SendTransactionWithOpts(ctx, tx, rpc.TransactionOpts{
			SkipPreflight:       false,
			PreflightCommitment: rpc.CommitmentProcessed,
			MinContextSlot:      &bh.Context.Slot,
		})
		if err == nil {
			break
		}
		log.Printf("Attempt %d: Failed to send transaction: %v", i+1, err)
		// Try to get more detailed error information
		if rpcErr, ok := err.(*jsonrpc.RPCError); ok {
			log.Printf("RPC Error details - Code: %d, Message: %s", rpcErr.Code, rpcErr.Message)
		}
		time.Sleep(time.Second * time.Duration(i+1))
	}
	if err != nil {
		return fmt.Errorf("failed to send transaction after retries: %w", err)
	}

	log.Printf("Create & Buy instructions sent - signature: %s", sig.String())
	return nil
}

func (c *RPCClient) AddBuyInstruction(mint solana.PublicKey, solAmount uint64) (*solana.GenericInstruction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	globalAccount, err := c.getGlobalAccount(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get global account: %w", err)
	}

	buyAmount, err := globalAccount.GetInitialBuyPrice(solAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate initial buy amount: %w", err)
	}
	slippagePercent := 10.0
	lamportsWithBuffer := uint64(float64(solAmount) * (1 + slippagePercent/100))

	program := solana.MustPublicKeyFromBase58(internal.PUMP_FUN_PROGRAM)
	bondingCurve, _, _ := programs.DeriveBondingCurve(mint, program)
	assocBondingCurve, _, _ := programs.DeriveAssociatedBondingCurve(mint, bondingCurve)
	assocUser, _, _ := programs.DeriveAssociatedTokenAccount(c.user.PublicKey(), mint)
	eventAuthority, _ := solana.PublicKeyFromBase58("Ce6TQqeHC9p8KetsN6JsjHK7UTZk7nasjjnr7XxXp9F1")
	creatorVault, _, _ := programs.DeriveCreatorVault(c.user.PublicKey(), program)

	buyIx := programs.NewBuyIx(
		buyAmount,
		lamportsWithBuffer,
		globalAccount.FeeRecipient,
		mint,
		assocBondingCurve,
		assocUser,
		c.user.PublicKey(),
		solana.SystemProgramID,
		solana.TokenProgramID,
		creatorVault,
		eventAuthority,
	)

	log.Printf("Buy instruction data - amount: %d, max_sol_cost: %d", buyAmount, lamportsWithBuffer)

	return buyIx, nil
}

func (c *RPCClient) AddCreateInstruction(metadata types.Metadata, metadataUri string) (*solana.GenericInstruction, error) {
	c.mint = solana.NewWallet()

	createIx := programs.NewCreateIx(
		c.mint.PublicKey(),
		c.user.PublicKey(),
		types.CreateData{
			Name:    metadata.Name,
			Symbol:  metadata.Symbol,
			Uri:     metadataUri,
			Creator: c.user.PublicKey(),
		},
	)

	return createIx, nil
}

func (c *RPCClient) getGlobalAccount(ctx context.Context) (*types.GlobalAccount, error) {
	programID := solana.MustPublicKeyFromBase58(internal.PUMP_FUN_PROGRAM)
	globalAccount, _, err := programs.DeriveGlobal(programID)
	if err != nil {
		return nil, fmt.Errorf("error deriving global account: %w", err)
	}

	var accountInfo *rpc.GetAccountInfoResult
	for i := 0; i < 3; i++ {
		accountInfo, err = c.rpcClient.GetAccountInfoWithOpts(ctx, globalAccount, &rpc.GetAccountInfoOpts{
			Commitment: rpc.CommitmentConfirmed,
		})
		if err == nil {
			break
		}
		log.Printf("Attempt %d: Failed to get account info: %v", i+1, err)
		time.Sleep(time.Second * time.Duration(i+1))
	}
	if err != nil {
		return nil, fmt.Errorf("error getting account info after retries: %w", err)
	}

	if accountInfo == nil {
		return nil, fmt.Errorf("global account not found")
	}

	rawData := accountInfo.Value.Data.GetBinary()
	if len(rawData) == 0 {
		return nil, fmt.Errorf("global account is empty")
	}

	var globalData types.GlobalAccount
	if err := borsh.Deserialize(&globalData, rawData); err != nil {
		log.Printf("Failed to deserialize global account data: %v", err)
		return nil, fmt.Errorf("error deserializing global account data: %w", err)
	}

	return &globalData, nil
}
