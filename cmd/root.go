/*
Copyright © 2024 Seednode <seednode@seedno.de>
*/

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

const (
	ReleaseVersion string = "1.2.0"
)

var (
	Bind            string
	Count           int
	Domain          string
	ErrorExit       bool
	Length          int
	Port            int
	Randomize       bool
	Scheme          string
	Timeout         time.Duration
	TimeoutInterval time.Duration
	URI             string
	Verbose         bool

	rootCmd = &cobra.Command{
		Use:   "send [file]...",
		Short: "Generates a one-off download link for one or more specified file(s).",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			switch {
			case Count < 0:
				return ErrInvalidCount
			case Length < 0:
				return ErrInvalidLength
			case Port < 1 || Port > 65535:
				return ErrInvalidPort
			case len(args) == 0 && !isFromPipe():
				return ErrNoFile
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return ServePage(args)
		},
	}
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Printf("%s | Error: %s\n", time.Now().Format(logDate), err)

		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&Bind, "bind", "b", "0.0.0.0", "address to bind to")
	rootCmd.Flags().IntVarP(&Count, "count", "c", 0, "number of times to serve the file(s)")
	rootCmd.Flags().StringVarP(&Domain, "domain", "d", "", "domain to use in returned urls")
	rootCmd.Flags().BoolVar(&ErrorExit, "error-exit", false, "shut down webserver on error, instead of just printing error")
	rootCmd.Flags().IntVarP(&Length, "length", "l", 6, "length of url slug and obfuscated filename(s)")
	rootCmd.Flags().IntVarP(&Port, "port", "p", 8080, "port to listen on")
	rootCmd.Flags().BoolVarP(&Randomize, "randomize", "r", false, "randomize filename(s)")
	rootCmd.Flags().StringVarP(&Scheme, "scheme", "s", "http", "scheme to use in returned urls")
	rootCmd.Flags().DurationVarP(&Timeout, "timeout", "t", 0, "shutdown after this length of time")
	rootCmd.Flags().DurationVar(&TimeoutInterval, "timeout-interval", time.Minute, "display remaining time in timeout every N seconds")
	rootCmd.Flags().StringVarP(&URI, "uri", "u", "", "full uri (overrides domain, scheme, and port)")
	rootCmd.Flags().BoolVarP(&Verbose, "verbose", "v", false, "log accessed files to stdout")

	rootCmd.Flags().SetInterspersed(true)

	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	rootCmd.SilenceErrors = true
	rootCmd.SetHelpCommand(&cobra.Command{
		Hidden: true,
	})

	rootCmd.SetVersionTemplate("send v{{.Version}}\n")
	rootCmd.Version = ReleaseVersion
}
