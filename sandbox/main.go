package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/viper"

	abciclient "github.com/tendermint/tendermint/abci/client"
	abci "github.com/tendermint/tendermint/abci/types"
	cfg "github.com/tendermint/tendermint/config"

	"github.com/tendermint/tendermint/node"

	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/libs/service"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "$HOME/.tendermint/config/config.toml", "Path to config.toml")
}

func main() {
	flag.Parse()
	config, err := readConfig(configFile)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := NewSandboxApp()
	tmNode, err := newTendermint(ctx, app, config)
	if err != nil {
		panic(err)
	}

	tmNode.Start(ctx)
	defer func() {
		cancel()
		tmNode.Wait()
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}

func readConfig(path string) (*cfg.Config, error) {
	config := cfg.DefaultValidatorConfig()
	config.RootDir = filepath.Dir(filepath.Dir(configFile))
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("viper failed to read config file: %w", err)
	}
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("viper failed to unmarshal config: %w", err)
	}
	if err := config.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("config is invalid: %w", err)
	}
	return config, nil
}

func newTendermint(ctx context.Context, app abci.Application, config *cfg.Config) (service.Service, error) {
	logger, err := log.NewDefaultLogger(log.LogFormatPlain, log.LogLevelInfo)
	if err != nil {
		return nil, err
	}
	return node.New(
		ctx,
		config,
		logger,
		abciclient.NewLocalClient(logger, app),
		nil)
}
