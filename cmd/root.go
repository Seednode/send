/*
Copyright © 2024 Seednode <seednode@seedno.de>
*/

package cmd

import (
	"errors"
	"time"

	"github.com/spf13/cobra"
)

const (
	ReleaseVersion string = "0.8.0"
)

var count uint32
var domain string
var length uint16
var port uint16
var randomize bool
var scheme string
var timeout time.Duration
var uri string
var verbose bool

var rootCmd = &cobra.Command{
	Use:   "send [file]...",
	Short: "Generates a one-off download link for one or more specified file(s).",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 && !isFromPipe() {
			err := errors.New("no file(s) specified and no data received from stdin")
			return err
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		ServePage(args)
	},
}

func Execute() {
	rootCmd.Execute()
}

func init() {
	rootCmd.Flags().Uint32VarP(&count, "count", "c", 0, "number of times to serve the file(s)")
	rootCmd.Flags().StringVarP(&domain, "domain", "d", "localhost", "domain to use in returned urls")
	rootCmd.Flags().Uint16VarP(&length, "length", "l", 6, "length of url slug and obfuscated filename(s)")
	rootCmd.Flags().Uint16VarP(&port, "port", "p", 8080, "port to listen on")
	rootCmd.Flags().BoolVarP(&randomize, "randomize", "r", false, "randomize filename(s)")
	rootCmd.Flags().StringVarP(&scheme, "scheme", "s", "http", "scheme to use in returned urls")
	rootCmd.Flags().DurationVarP(&timeout, "timeout", "t", 0, "shutdown after this length of time")
	rootCmd.Flags().StringVarP(&uri, "uri", "u", "", "full uri (overrides domain, scheme, and port)")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "log accessed files to stdout")

	rootCmd.Flags().SetInterspersed(true)

	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	rootCmd.SilenceErrors = true
	rootCmd.SetHelpCommand(&cobra.Command{
		Hidden: true,
	})

	rootCmd.SetVersionTemplate("send v{{.Version}}\n")
	rootCmd.Version = ReleaseVersion
}
