package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/rs/zerolog"

	"github.com/warden-protocol/discord-faucet/pkg/config"
	"github.com/warden-protocol/discord-faucet/pkg/discord"
)

const (
	serverTimeout = 10
)

func main() {
	logger := log.New(
		log.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339},
	).Level(log.InfoLevel).With().Timestamp().Logger()

	// load configuration
	config, err := config.LoadConfig()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// init discordbot
	discordBot, err := discord.InitDiscord(config)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize discord bot")
	}

	discordBot.Session.AddHandler(discordBot.MessageCreate)

	discordBot.Session.Identify.Intents = discordgo.IntentsGuildMessages

	if err = discordBot.Session.Open(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to open connection")
	}
	logger.Info().Msg("Bot is now running")
	go discordBot.StartPurgeRoutine()

	http.Handle("/metrics", promhttp.Handler())

	logger.Info().Msgf("starting metrics server on port %s", config.Port)
	server := &http.Server{
		Addr:              ":" + config.Port,
		ReadHeaderTimeout: serverTimeout * time.Second,
	}

	if err = server.ListenAndServe(); err != nil {
		logger.Fatal().Err(err).Msgf("error starting server: %v", err)
	}

	// Wait here until CTRL-C or other term signal is received.
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
