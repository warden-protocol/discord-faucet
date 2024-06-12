package discord

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/caarlos0/env/v9"
	"github.com/rs/zerolog"

	"github.com/warden-protocol/discord-faucet/pkg/faucet"
)

const (
	defaultPurgeInterval = 10
)

type Discord struct {
	Session       *discordgo.Session
	Token         string `env:"TOKEN" envDefault:""`
	PurgeInterval time.Duration
	Requests      map[string]time.Time
	Faucet        faucet.Faucet
	logger        zerolog.Logger
	*sync.Mutex
}

func InitDiscord() (Discord, error) {
	var err error

	d := Discord{
		Mutex: &sync.Mutex{},
	}
	if err = env.Parse(&d); err != nil {
		return Discord{}, err
	}

	d.logger = zerolog.New(
		zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339},
	).Level(zerolog.InfoLevel).With().Timestamp().Caller().Logger()

	if d.Token == "" {
		return Discord{}, fmt.Errorf("missing discord token")
	}

	d.logger.Info().Msg("initialising discord connection")
	d.Session, err = discordgo.New("Bot " + d.Token)
	if err != nil {
		return Discord{}, err
	}

	d.Requests = make(map[string]time.Time)

	d.logger.Info().Msg("initialising faucet")
	d.Faucet, err = faucet.InitFaucet()
	if err != nil {
		return Discord{}, err
	}
	d.Faucet.Logger = d.logger

	interval := os.Getenv("PURGE_INTERVAL")
	if interval == "" {
		d.PurgeInterval = defaultPurgeInterval * time.Second
	} else {
		d.PurgeInterval, err = time.ParseDuration(interval)
		if err != nil {
			return Discord{}, err
		}
	}

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
	reqCount.Inc()
	addr := strings.TrimSpace(strings.Split(m.Content, "$request ")[1])
	if addr == "" {
		reqBad.Inc()
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
			reqDenied.Inc()
			return
		}
	}

	var returnMsg string
	tx, err := d.Faucet.Send(addr, 1)
	if err != nil {
		reqFailed.Inc()
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
		usersCooldown.Inc()
	}
	_, err = d.Session.ChannelMessageSend(m.ChannelID, returnMsg)
	if err != nil {
		d.logger.Error().Err(err).Msgf("failed to send message")
		return
	}
}

func (d *Discord) purgeExpiredEntries() {
	d.Lock()
	defer d.Unlock()

	now := time.Now()
	for k, v := range d.Requests {
		diff := now.Sub(v)
		if diff > d.Faucet.Cooldown {
			usersCooldown.Sub(1)
			delete(d.Requests, k)
			d.logger.Info().Msgf("purged entry for key: %s", k)
		}
	}
	for k, v := range d.Faucet.Requests {
		diff := now.Sub(v)
		if diff > d.Faucet.Cooldown {
			delete(d.Faucet.Requests, k)
			d.logger.Info().Msgf("purged entry for key: %s", k)
		}
	}
}

//nolint:gosimple // need a while loop for this
func (d *Discord) StartPurgeRoutine() {
	ticker := time.NewTicker(d.PurgeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			d.purgeExpiredEntries()
		}
	}
}
