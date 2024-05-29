package faucet

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/caarlos0/env/v9"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/rs/zerolog"
)

type Faucet struct {
	Cooldown    time.Duration
	Mnemonic    string `env:"MNEMONIC" envDefault:""`
	Node        string `env:"NODE" envDefault:"https://rpc.buenavista.wardenprotocol.org:443"`
	ChainID     string `env:"CHAIN_ID" envDefault:"buenavista-1"`
	CliName     string `env:"CLI_NAME" envDefault:"wardend"`
	AccountName string `env:"ACCOUNT_NAME" envDefault:"faucet"`
	Denom       string `env:"DENOM" envDefault:"uward"`
	Amount      string `env:"AMOUNT" envDefault:"10000000"`
	Fees        string `env:"FEES" envDefault:"25uward"`
	TXRetry     int    `env:"TX_RETRY" envDefault:"10"`
	Requests    map[string]time.Time
	Logger      zerolog.Logger
}

const (
	waitTime = 2
)

type Out struct {
	Stdout []byte
	Stderr []byte
}

func envOrDefault(key, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	return v
}

func execute(cmdString string) (Out, error) {
	// Create the command
	cmd := exec.Command("sh", "-c", cmdString)

	// Get the output pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return Out{}, fmt.Errorf("error getting stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return Out{}, fmt.Errorf("error getting stdout pipe: %w", err)
	}

	// Start the command
	if err = cmd.Start(); err != nil {
		return Out{}, fmt.Errorf("error getting stdout pipe: %w", err)
	}

	// Read the output
	output, err := io.ReadAll(stdout)
	if err != nil {
		return Out{}, fmt.Errorf("error getting stdout pipe: %w", err)
	}
	errOutput, err := io.ReadAll(stderr)
	if err != nil {
		return Out{}, fmt.Errorf("error getting stdout pipe: %w", err)
	}

	// Wait for the command to finish
	if err = cmd.Wait(); err != nil {
		return Out{}, fmt.Errorf("error getting stdout pipe: %w", err)
	}

	return Out{Stdout: output, Stderr: errOutput}, nil
}

func (f *Faucet) setupNewAccount() error {
	cmd := strings.Join([]string{
		"echo",
		f.Mnemonic,
		"|",
		f.CliName,
		"keys",
		"--keyring-backend",
		"test",
		"add",
		f.AccountName,
		"--recover",
	}, " ")

	_, err := execute(cmd)
	if err != nil {
		return err
	}
	return nil
}

func validAddress(addr string) error {
	pref, _, err := bech32.DecodeAndConvert(addr)
	if err != nil {
		return fmt.Errorf("invalid address: %w", err)
	}
	if pref != "warden" {
		return fmt.Errorf("invalid address prefix: %s", pref)
	}
	return nil
}

func InitFaucet() (Faucet, error) {
	var err error

	f := Faucet{}
	if err = env.Parse(&f); err != nil {
		return Faucet{}, err
	}

	f.Cooldown, err = time.ParseDuration(envOrDefault("COOLDOWN", "12h"))
	if err != nil {
		return Faucet{}, err
	}

	if f.Mnemonic == "" {
		return Faucet{}, fmt.Errorf("missing mnemonic")
	}

	if err = f.setupNewAccount(); err != nil {
		return Faucet{}, err
	}

	f.Requests = make(map[string]time.Time)

	return f, nil
}

func (f *Faucet) Send(addr string, retry int) (string, error) {
	if err := validAddress(addr); err != nil {
		return "", err
	}

	if t, found := f.Requests[addr]; found {
		now := time.Now()
		diff := now.Sub(t)
		if diff < f.Cooldown {
			waitTime := f.Cooldown - diff
			return "", fmt.Errorf("address %s needs to wait for %v", addr, waitTime)
		}
	}

	f.Logger.Info().Msgf("sending %s%s to %v", f.Amount, f.Denom, addr)

	amount := f.Amount + f.Denom

	cmd := strings.Join([]string{
		f.CliName,
		"tx",
		"bank",
		"send",
		f.AccountName,
		addr,
		amount,
		"--yes",
		"--keyring-backend",
		"test",
		"--chain-id",
		f.ChainID,
		"--node",
		f.Node,
		"--gas-prices",
		f.Fees,
		"-o",
		"json",
	}, " ")

	out, err := execute(cmd)
	if err != nil {
		return "", err
	}

	var result struct {
		Code   int    `json:"code"`
		TxHash string `json:"txhash"`
	}

	if err = json.Unmarshal(out.Stdout, &result); err != nil {
		return "", fmt.Errorf("error unmarshalling tx result: %w", err)
	}
	if result.Code == 32 && retry < f.TXRetry {
		f.Logger.Info().Msgf(
			"tx failed with code %d for address %s, retrying (%d/%d)",
			result.Code,
			addr,
			retry,
			f.TXRetry,
		)
		time.Sleep(waitTime * time.Second)
		return f.Send(addr, retry+1)
	}
	if result.Code != 0 {
		return "", fmt.Errorf(
			"tx failed with code %d for address %s",
			result.Code,
			addr,
		)
	}

	return result.TxHash, nil
}
