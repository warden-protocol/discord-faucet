package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	log "github.com/rs/zerolog"

	"github.com/warden-protocol/discord-faucet/pkg/discord"
)

func main() {
	logger := log.New(
		log.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339},
	).Level(log.TraceLevel).With().Timestamp().Caller().Logger()

	discordBot, err := discord.InitDiscord()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize discord bot")
	}

	discordBot.Session.AddHandler(discordBot.MessageCreate)

	// // In this example, we only care about receiving message events.
	discordBot.Session.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	if err = discordBot.Session.Open(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to open connection")
	}

	logger.Info().Msg("Bot is now running")
	// Wait here until CTRL-C or other term signal is received.
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	if err = discordBot.Session.Close(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to close connection")
	}
}
