package main

import (
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
		nil,
		// Meteora DLMM
		[]string{"LBUZKhRxPF3XUpBCjp4YzTKgLccjZhTSDM9YuVaPwxo"},
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
	log.Info().Any("resp", resp).Msg("subscription response received")
}
