package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/caarlos0/env/v10"
	"github.com/spf13/viper"
)

var errConfig = errors.New("config error")

func configError(msg string) error {
	return fmt.Errorf("%w: %s", errConfig, msg)
}

type Config struct {
	Port          string `env:"PORT" envDefault:"8081" mapstructure:"PORT"`
	EnvFile       string `env:"ENV_FILE" envDefault:""`
	Token         string `env:"TOKEN" envDefault:"" mapstructure:"TOKEN"`
	PurgeInterval string `env:"PURGE_INTERVAL" envDefault:"10s" mapstructure:"PURGE_INTERVAL"`
	Mnemonic      string `env:"MNEMONIC" envDefault:"" mapstructure:"MNEMONIC"`
	Node          string `env:"NODE" envDefault:"https://rpc.chiado.wardenprotocol.org:443" mapstructure:"NODE"`
	ChainID       string `env:"CHAIN_ID" envDefault:"chiado_10010-1" mapstructure:"CHAIN_ID"`
	CliName       string `env:"CLI_NAME" envDefault:"wardend" mapstructure:"CLI_NAME"`
	AccountName   string `env:"ACCOUNT_NAME" envDefault:"faucet" mapstructure:"ACCOUNT_NAME"`
	Denom         string `env:"DENOM" envDefault:"award" mapstructure:"DENOM"`
	Amount        string `env:"AMOUNT" envDefault:"10" mapstructure:"AMOUNT"`
	Decimal       int    `env:"DECIMAL" envDefault:"18" mapstructure:"DECIMAL"`
	Fees          string `env:"FEES" envDefault:"25000000000000award" mapstructure:"FEES"`
	TXRetry       int    `env:"TX_RETRY" envDefault:"10" mapstructure:"TX_RETRY"`
	CoolDown      string `env:"COOLDOWN" envDefault:"10s" mapstructure:"COOLDOWN"`
}

func LoadConfig() (Config, error) {
	cfg := Config{}
	var err error

	// setDefaults(*cfg)

	if err = env.Parse(&cfg); err != nil {
		return Config{}, configError(err.Error())
	}

	if cfg.EnvFile != "" {
		if err = loadConfigFile(&cfg); err != nil {
			return Config{}, configError(err.Error())
		}
	}
	return cfg, nil
}

func loadConfigFile(cfg *Config) error {
	var err error

	// parse config file params
	// Extract the directory
	dir := filepath.Dir(cfg.EnvFile) + "/"

	// Extract the base name (filename without directory)
	base := filepath.Base(cfg.EnvFile)

	// Split the base name into name and extension
	name := strings.TrimSuffix(base, filepath.Ext(base))
	ext := strings.TrimPrefix(filepath.Ext(base), ".")

	viper.AddConfigPath(dir)
	viper.SetConfigName(name)
	viper.SetConfigType(ext)

	viper.AutomaticEnv()
	err = viper.ReadInConfig()
	if err != nil {
		return err
	}

	if err = viper.Unmarshal(&cfg); err != nil {
		return err
	}
	return nil
}
