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
	// Version number for built binaries and Docker image releases
	ReleaseVersion string = "2.4.0"
)

var (
	// The IP address on which send will listen
	Bind string

	// The number of times to serve selected file(s) before shutting down
	Count int

	// Exit on error, instead of just printing the error
	ErrorExit bool

	// The length of randomly generated slugs and filenames
	Length int

	// The port on which send will listen
	Port int

	// Register http/pprof handlers
	Profile bool

	// Randomize filenames in URLs
	Randomize bool

	// The length of time after which send will shut down
	Timeout time.Duration

	// How often to display remaining time before shutdown, when timeout is enabled
	TimeoutInterval time.Duration

	// Value to be used instead of http://<bind>:<port> in returned links
	URL string

	rootCmd = &cobra.Command{
		Use:   "send [file]...",
		Short: "Generates a one-off download link for one or more specified files.",
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
		fmt.Println(err)

		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&Bind, "bind", "b", "0.0.0.0", "address to bind to")
	rootCmd.Flags().IntVarP(&Count, "count", "c", 0, "number of times to serve files")
	rootCmd.Flags().BoolVarP(&ErrorExit, "exit", "e", false, "shut down webserver on error, instead of just printing error")
	rootCmd.Flags().IntVarP(&Length, "length", "l", 6, "length of url slug and obfuscated filenames")
	rootCmd.Flags().IntVarP(&Port, "port", "p", 8080, "port to listen on")
	rootCmd.Flags().BoolVar(&Profile, "profile", false, "register net/http/pprof handlers")
	rootCmd.Flags().BoolVarP(&Randomize, "randomize", "r", false, "randomize filenames")
	rootCmd.Flags().DurationVarP(&Timeout, "timeout", "t", 0, "shutdown after this length of time")
	rootCmd.Flags().DurationVarP(&TimeoutInterval, "interval", "i", time.Minute, "display remaining time in timeout at this interval")
	rootCmd.Flags().StringVarP(&URL, "url", "u", "", "use this value instead of http://<bind>:<port> in returned URLs")

	rootCmd.Flags().SetInterspersed(true)

	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true
	rootCmd.SetHelpCommand(&cobra.Command{
		Hidden: true,
	})

	rootCmd.SetVersionTemplate("send v{{.Version}}\n")
	rootCmd.Version = ReleaseVersion
}
