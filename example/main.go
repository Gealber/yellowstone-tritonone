package main

import (
	"os"
	"time"

	"github.com/Gealber/base58"
	grpcClt "github.com/Gealber/yellowstone-tritonone/client"
	"github.com/Gealber/yellowstone-tritonone/proto"
	"github.com/gagliardetto/solana-go"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

var (
	pids = []string{
		// "LBUZKhRxPF3XUpBCjp4YzTKgLccjZhTSDM9YuVaPwxo",
		// solana.SystemProgramID.String(),
		// "vnt1u7PzorND5JjweFWmDawKe2hLWoTwHU6QKz6XX98",
		// "4GCrA5ATXg5GixzjW9ZZTXNh5fjWPvhanN5p7YEgxwuA",
		"Vote111111111111111111111111111111111111111",
		// "Czfq3xZZDmsdGdUyrNLtRhGc47cXcZtLG4crryfu44zE",
		// "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
		// "TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb",
		// "AddressLookupTab1e1111111111111111111111111",
		// "6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P",
	}
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading .env file")
	}

	if len(os.Args) > 1 {
		os.Setenv("GRPC_TOKEN", os.Args[1])
	}

	commitment := proto.CommitmentLevel_PROCESSED

	clt, err := grpcClt.New(
		// []string{"5rCf1DM8LjKTw4YqhnoLcngyZYeNnQqztScTogYHAS6"},
		nil,
		// Meteora DLMM
		nil,
		pids,
		// []string{"EN2arTWbbUcyqXDE8mTykd2sycQZEuBnHKTP9KjuV9Pw"},
		// []string{"Vote111111111111111111111111111111111111111"},
		// []string{"Ed25519SigVerify111111111111111111111111111", "vnt1u7PzorND5JjweFWmDawKe2hLWoTwHU6QKz6XX98"},
		// []string{"Vote111111111111111111111111111111111111111"},
		false, // blockSub
		false, // slotSub
		processSub,
		&commitment,
	)
	if err != nil {
		panic(err)
	}

	err = clt.Run()
	if err != nil {
		panic(err)
	}
}

func processSub(resp *proto.SubscribeUpdate) {
	slotUpd := resp.GetSlot()
	if slotUpd != nil {
		log.Info().
			Uint64("slot", slotUpd.Slot).
			Int64("ts", time.Now().Unix()).
			Msg("Slot")
		return
	}

	blkUpd := resp.GetBlock()
	if blkUpd != nil {
		log.Info().
			Str("blk_hash", blkUpd.Blockhash).
			Uint64("slot", blkUpd.Slot).
			Int64("slot_ts", blkUpd.BlockTime.Timestamp).
			Int64("ts", time.Now().Unix()).
			Msg("Block")
		return
	}

	upd := resp.GetTransaction()
	if upd != nil {
		sig := upd.GetTransaction().Signature
		sigStr := base58.Encode64([64]byte(sig))
		log.Info().Str("signature", sigStr).Msg("subscription response received")
		return
	}

	accUpd := resp.GetAccount()
	if accUpd != nil {
		acc := accUpd.GetAccount()
		pk := solana.PublicKeyFromBytes(acc.Pubkey)
		log.Info().
			Uint64("slot", accUpd.GetSlot()).
			Str("pk", pk.String()).
			Uint64("lamports", acc.Lamports).
			Msg("subscription response received>>>>>>>>>>>>>>>>>>>>>>>>>")
		return
	}
}
