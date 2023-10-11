package cmd

import (
	"fmt"

	"github.com/pkg/profile"
	"github.com/spf13/cobra"
)

func wrapProfiling(f func([]string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true

		if !cmd.Flags().Changed("pprof") {
			return f(args)
		}

		opts := []func(*profile.Profile){profile.ProfilePath(".")}

		if b, _ := cmd.Flags().GetBool("debug"); !b {
			opts = append(opts, profile.Quiet)
		}

		switch s, _ := cmd.Flags().GetString("pprof"); s {
		case "cpu":
			opts = append(opts, profile.CPUProfile)
		case "clock":
			opts = append(opts, profile.ClockProfile)
		case "mem_heap":
			opts = append(opts, profile.MemProfileHeap)
		case "mem_allocs":
			opts = append(opts, profile.MemProfileAllocs)
		case "trace":
			opts = append(opts, profile.TraceProfile)
		case "goroutine":
			opts = append(opts, profile.GoroutineProfile)
		case "mutex":
			opts = append(opts, profile.MutexProfile)
		case "block":
			opts = append(opts, profile.BlockProfile)
		case "thread_creation":
			opts = append(opts, profile.ThreadcreationProfile)
		default:
			return fmt.Errorf("unknown profile type %s", s)
		}

		defer profile.Start(opts...).Stop()

		return f(args)
	}
}
