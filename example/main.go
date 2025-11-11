package main

import (
	"github.com/Gealber/base58"
	grpcClt "github.com/Gealber/yellowstone-tritonone/client"
	"github.com/Gealber/yellowstone-tritonone/proto"
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
		// nil,
		// []string{solana.SystemProgramID.String()},
		// Meteora DLMM
		[]string{"LBUZKhRxPF3XUpBCjp4YzTKgLccjZhTSDM9YuVaPwxo"},
		nil,
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
		pk := base58.Encode32([32]byte(acc.Pubkey))
		log.Info().Str("pk", pk).Msg("subscription response received")
		return
	}
}
