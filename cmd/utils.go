package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/styrainc/regal/pkg/config"
)

// configFileParams supports extracting the config file path from various command
// param types. This allows readUserConfig to be shared.
type configFileParams interface {
	getConfigFile() string
}

func readUserConfig(params configFileParams, regalDir *os.File) (userConfig *os.File, err error) {
	if cfgFile := params.getConfigFile(); cfgFile != "" {
		userConfig, err = os.Open(cfgFile)
		if err != nil {
			return nil, fmt.Errorf("failed to open config file %w", err)
		}
	} else {
		searchPath, _ := os.Getwd()
		if regalDir != nil {
			searchPath = regalDir.Name()
		}

		if searchPath != "" {
			userConfig, err = config.FindConfig(searchPath)
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
