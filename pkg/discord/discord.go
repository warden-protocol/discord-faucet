package discord

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/caarlos0/env/v9"
	"github.com/rs/zerolog"

	"github.com/warden-protocol/discord-faucet/pkg/faucet"
)

type Discord struct {
	Session  *discordgo.Session
	Token    string `env:"TOKEN" envDefault:""`
	Requests map[string]time.Time
	Faucet   faucet.Faucet
	logger   zerolog.Logger
}

func InitDiscord() (Discord, error) {
	var err error

	d := Discord{}
	if err = env.Parse(&d); err != nil {
		return Discord{}, err
	}

	d.logger = zerolog.New(
		zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339},
	).Level(zerolog.InfoLevel).With().Timestamp().Caller().Logger()

	if d.Token == "" {
		return Discord{}, fmt.Errorf("missing discord token")
	}
	d.Session, err = discordgo.New("Bot " + d.Token)
	if err != nil {
		return Discord{}, err
	}

	d.Requests = make(map[string]time.Time)

	d.Faucet, err = faucet.InitFaucet()
	if err != nil {
		return Discord{}, err
	}
	d.Faucet.Logger = d.logger

	return d, nil
}

func (d *Discord) MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.Contains(m.Content, "$request") {
		go d.requestFunds(m)
	}
}

func (d *Discord) requestFunds(m *discordgo.MessageCreate) {
	d.logger.Info().Msgf("user %s requested funds to %s", m.Author, m.Content)
	addr := strings.Split(m.Content, "$request ")[1]
	if addr == "" {
		d.logger.Error().Msgf("missing address for user %s", m.Author)
		return
	}

	if user, found := d.Requests[m.Author.Username]; found {
		now := time.Now()
		diff := now.Sub(user)
		if diff < d.Faucet.Cooldown {
			waitTime := d.Faucet.Cooldown - diff
			_, err := d.Session.ChannelMessageSend(
				m.ChannelID,
				fmt.Sprintf(":red_circle: user %s needs to wait for %v",
					m.Author.Username,
					waitTime,
				))
			if err != nil {
				d.logger.Error().Err(err).Msgf("failed to send message")
				return
			}
			return
		}
	}

	var returnMsg string
	tx, err := d.Faucet.Send(addr)
	if err != nil {
		d.logger.Error().Err(err).Msgf("failed to send funds to %s", addr)
		returnMsg = fmt.Sprintf(":red_circle: %s", err)
	} else {
		returnMsg = fmt.Sprintf(
			":white_check_mark: 10 WARD sent to address %s \n %s",
			addr,
			fmt.Sprintf("https://testnet.warden.explorers.guru/transaction/%s", tx),
		)
		d.Requests[m.Author.Username] = time.Now()
		d.Faucet.Requests[addr] = time.Now()
	}
	_, err = d.Session.ChannelMessageSend(m.ChannelID, returnMsg)
	if err != nil {
		d.logger.Error().Err(err).Msgf("failed to send message")
		return
	}
}
