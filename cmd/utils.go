package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/styrainc/regal/pkg/config"
)

// configFileParams supports extracting the config file path from various command
// param types. This allows readUserConfig to be shared.
type configFileParams interface {
	getConfigFile() string
}

func readUserConfig(params configFileParams, searchPath string) (userConfig *os.File, err error) {
	if cfgFile := params.getConfigFile(); cfgFile != "" {
		userConfig, err = os.Open(cfgFile)
		if err != nil {
			return nil, fmt.Errorf("failed to open config file %w", err)
		}
	} else {
		if searchPath == "" {
			searchPath, _ = os.Getwd()
		}

		userConfig, err = config.FindConfig(searchPath)
	}

	// if there is no config found, attempt to load the user's global config if
	// it exists
	if err != nil {
		globalConfigDir := config.GlobalConfigDir(false)
		if globalConfigDir != "" {
			globalConfigFile := filepath.Join(globalConfigDir, "config.yaml")

			userConfig, err = os.Open(globalConfigFile)
			if err != nil {
				return nil, fmt.Errorf("failed to open global config file %w", err)
			}
		}
	}

	return userConfig, err //nolint:wrapcheck
}

// timeoutParams supports extracting the timeout duration from various command
// param types. This allows getLinterContext to be shared.
type timeoutParams interface {
	getTimeout() time.Duration
}

func getLinterContext(params timeoutParams) (context.Context, func()) {
	ctx := context.Background()

	cancel := func() {}

	if to := params.getTimeout(); to != 0 {
		ctx, cancel = context.WithTimeout(ctx, to)
	}

	return ctx, cancel
}
