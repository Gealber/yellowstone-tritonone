package main

import (
	"time"

	"github.com/Gealber/base58"
	grpcClt "github.com/Gealber/yellowstone-tritonone/client"
	"github.com/Gealber/yellowstone-tritonone/proto"
	"github.com/gagliardetto/solana-go"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading .env file")
	}

	clt, err := grpcClt.New(
		nil,
		// Meteora DLMM
		// []string{"LBUZKhRxPF3XUpBCjp4YzTKgLccjZhTSDM9YuVaPwxo"},
		nil,
		// []string{"Vote111111111111111111111111111111111111111"},
		nil,
		false,
		true,
		processSub,
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
		log.Info().Uint64("slot", accUpd.GetSlot()).Str("pk", pk.String()).Msg("subscription response received>>>>>>>>>>>>>>>>>>>>>>>>>")
		return
	}
}
